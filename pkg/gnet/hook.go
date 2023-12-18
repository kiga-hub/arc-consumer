package gnet

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/ringbuffer"
)

// OnInitComplete server 初始化完成之后调用
func (cs *Server) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	cs.logger.Infow("listening",
		"addr", srv.Addr.String(),
		"cores", srv.Multicore,
		"loops", srv.NumEventLoop,
	)
	return
}

// OnOpened 连接被打开的时候调用
func (cs *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	cs.metrics.connections.Inc()
	cs.logger.Infow("opened", "addr", c.RemoteAddr().String())
	return
}

// OnClosed 连接被关闭的时候调用
func (cs *Server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	cs.metrics.connections.Dec()
	cs.logger.Infow("closed", "addr", c.RemoteAddr().String())

	cs.sensors.Range(func(key, value interface{}) bool {
		sensor := value.(*Sensor)
		if sensor.addr != c.RemoteAddr().String() {
			return true
		}

		// if cs.config.NoticeEnable {
		// 	dev := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		// 		Host:     cs.config.DeviceHost,
		// 		BasePath: "/",
		// 		Schemes:  []string{"http"},
		// 	})
		// 	res, err := dev.Collector.CollectorDataNotice(&collector.CollectorDataNoticeParams{
		// 		Notice: &models.CollectorNoticeRequest{
		// 			SensorID: int64(key.(uint64)),
		// 			Action:   1,
		// 		},
		// 		Context: context.Background(),
		// 	})
		// 	if err != nil {
		// 		cs.logger.Errorw(err.Error())
		// 	}
		// 	cs.logger.Info(res)
		// }

		if !sensor.holding {
			cs.sensors.Delete(key)
		}
		return true
	})
	return
}

// OnShutdown 所有event-loop 连接关闭后被调用
func (cs *Server) OnShutdown(server gnet.Server) {
	cs.logger.Infow("shut down")
	cs.frameChans.Range(func(key, value interface{}) bool {
		v := value.(chan *Package)
		close(v)
		cs.frameChans.Delete(key)
		return true
	})
}

// React 注册一个React事件，业务逻辑代码部分，接收完整的一个Frame时被调用
func (cs *Server) React(data []byte, c gnet.Conn) ([]byte, gnet.Action) {
	addr := c.RemoteAddr().String()
	/*
		这里不能使用gnet的工作池统一处理所有的包，会出现包顺序错乱的问题
	*/

	// 获取传感器ID
	var sensorid uint64
	for _, b := range data[23:29] {
		sensorid <<= 8
		sensorid += uint64(b)
	}

	// 获取包号
	seq := int64(binary.BigEndian.Uint64(data[9:17]))

	var isResp bool
	if data[37]&0x20 != 0 {
		isResp = true
	}
	if seq <= 2 && !isResp {
		cs.logger.Infow("drop packet", "addr", addr, "sensor", fmt.Sprintf("%012X", sensorid), "seq", seq)
		return nil, gnet.None
	}
	if data[37]&0x40 != 0 {
		cs.logger.Infow("end packet", "addr", addr, "sensor", fmt.Sprintf("%012X", sensorid), "seq", seq)
	}

	// 获取传感器信息
	var sensor *Sensor
	value, ok := cs.sensors.Load(sensorid)
	if !ok {
		// 初始化新的传感器管理结构
		sensor = &Sensor{
			id:              sensorid,
			sid:             fmt.Sprintf("%012X", sensorid),
			addr:            c.RemoteAddr().String(),
			logger:          cs.logger,
			offsetThreshold: cs.config.OffsetThreshold * 1e3,
		}

		// 判断重发，分配环形队列
		if cs.config.ResendEnable {
			sensor.buff = ringbuffer.New(1024 * 2048)
		}

		// 保存传感器结构
		cs.sensors.Store(sensorid, sensor)
	} else {
		sensor = value.(*Sensor)
	}

	// 判断断开连接是否清除缓存结构
	if isResp && !sensor.holding {
		cs.logger.Infow("struct holding", "addr", strings.Split(addr, ":")[0], "sensor", fmt.Sprintf("%012X", sensorid))
		sensor.holding = true
	}

	// 统计信息
	sensor.lasttime = time.Now().UnixNano() / 1e3
	sensor.sampleCount++
	if sensor.sampleCount > 1000 {
		sensor.sample = true
		sensor.sampleCount = 0
	}
	cs.metrics.packets.WithLabelValues(sensor.sid).Inc()
	cs.metrics.bytes.WithLabelValues(sensor.sid, "all").Add(float64(len(data) * 8))

	// 判断过滤重复包
	if isResp && seq == sensor.lastSequence {
		var resp = make([]byte, 8)
		binary.BigEndian.PutUint64(resp, uint64(seq))
		cs.logger.Infow("repeat package", "addr", strings.Split(addr, ":")[0], "sensor", fmt.Sprintf("%012X", sensorid), "seq", seq)
		return resp, gnet.None
	}

	// 判断执行重发机制
	if cs.config.ResendEnable {
		// 判断重发
		if req, ok := cs.ResendHandle(seq, sensor, data); ok {
			return req, gnet.None
		}
	}

	// 提交处理当前包
	cs.ToHandle(sensor, data, sensor.lasttime)

	// 检查处理缓存内已经连续包号的数据包
	if cs.config.ResendEnable {
		cs.BufferToHandle(sensor, true)
	}

	// 判断包号应答
	if isResp {
		var resp = make([]byte, 8)
		binary.BigEndian.PutUint64(resp, uint64(seq))
		cs.logger.Debugw("ack package", "addr", strings.Split(addr, ":")[0], "sensor", fmt.Sprintf("%012X", sensorid), "seq", seq)
		return resp, gnet.None
	}

	return nil, gnet.None
}

// Tick -
func (cs *Server) Tick() (time.Duration, gnet.Action) {
	return 0, gnet.None
}
