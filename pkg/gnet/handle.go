package gnet

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/kiga-hub/arc/protocols"
)

// VSamplerateTable - 振动采样率对照表
var VSamplerateTable = []float64{3200, 1600, 800, 400, 200, 100, 50, 25, 12.5, 6.25, 3.13, 1.56, 0.78, 0.39, 0.2, 0.1}

// Package - 处理包结构
type Package struct {
	Sensor   *Sensor // 传感器
	Realtime int64
	Data     []byte // 数据
}

// 获取设备型号
func getModel(flag [3]byte) string {
	model := ""

	// 传感器探头标志量
	if flag[1]&0x01 != 0 {
		model += "A"
	}
	if flag[1]&0x02 != 0 {
		model += "V"
	}
	if flag[1]&0x04 != 0 {
		model += "T"
	}

	if model != "" {
		model += "-"
	}

	// 有线、无线标志量
	if flag[0]&0x02 != 0 {
		model += "W"
	} else {
		model += "C"
	}

	model = fmt.Sprintf("%s%02d", model, flag[0]>>2)

	return model
}

// 包时间对齐、统计
func (cs *Server) decodePackage(frameBuff *protocols.Frame, data []byte, sensor *Sensor, realtime int64) (int64, error) {
	// 解Frame数据包
	/*
		frameBuff := protocols.NewDefaultFrame()
		if err := frameBuff.Decode(data); err != nil {
			return 0, err
		}
	*/

	// 从Frame包获取音频数据段(获取Frame包对齐时间是根据音频时间计算的)
	// 统计各个数据段大小
	var err error
	var sa *protocols.SegmentArc
	for _, stype := range frameBuff.DataGroup.STypes {
		switch stype {
		case protocols.STypeArc:
			sa, err = frameBuff.DataGroup.GetArcSegment()
			if err != nil {
				return 0, err
			}

			// 数据段长度减头信息大小为数据实际长度
			cs.metrics.bytes.WithLabelValues(sensor.sid, "arc").Add(float64(sa.Size() - 5))
		}
	}
	sensor.sample = false

	// Frame包内必须有音频数据段
	if sa == nil {
		return 0, fmt.Errorf("%012X package not find audio data", sensor.id)
	}

	// 根据音频数据，计算数据时间（微秒）
	dataSize := int64(len(sa.Data))
	return dataSize, nil
}

// 获取goroutine id
func (cs *Server) getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		cs.logger.Errorf("cannot get goroutine id %v", err)
	}
	return n
}

// 从管道获取package结构，包处理
func (cs *Server) handlePackage(frameBuff *protocols.Frame, pkg *Package) {
	cs.metrics.queuePkts.WithLabelValues(pkg.Sensor.sid).Dec()

	// 集群同步上报传感器编号
	if cs.kvCache != nil {
		if err := cs.kvCache.Sync(pkg.Sensor.id); err != nil {
			cs.logger.Errorw(err.Error(), "sensor", pkg.Sensor.sid)
		}
	}

	// 数据包tcp转发
	if cs.proxy != nil && !cs.config.ProxyTimealign {
		if _, err := cs.proxy.Write(pkg.Sensor.id, pkg.Sensor.sid, pkg.Data); err != nil {
			cs.logger.Warnw(err.Error(), "sensor", pkg.Sensor.sid)
		}
	}

	/*
	 *  Frame数据包实时转发到其他服务(加时间戳)，目前方式有2种：
	 *	1 - websocket
	 *	2 - grpc
	 *
	 *	包数据修改了两个地方:
	 *	1 - 对齐时间替换包号
	 * 	2 - 置对齐时间连续标志位
	 *
	 */

	// 解包
	if err := frameBuff.Decode(pkg.Data); err != nil {
		cs.logger.Errorw(err.Error(), "sensor", pkg.Sensor.sid)
		return
	}

	// 时间对齐，统计检查
	_, err := cs.decodePackage(frameBuff, pkg.Data, pkg.Sensor, pkg.Realtime)
	if err != nil {
		cs.logger.Errorw(err.Error())
		return
	}

	cs.tmap.Store(pkg.Sensor.id, frameBuff.Timestamp)

	// 数据包tcp转发
	if cs.proxy != nil && cs.config.ProxyTimealign {
		if _, err := cs.proxy.Write(pkg.Sensor.id, pkg.Sensor.sid, pkg.Data); err != nil {
			cs.logger.Warnw(err.Error(), "sensor", pkg.Sensor.sid)
		}
	}

	// 数据包websocket转发
	if cs.httpproxy != nil {
		if _, err := cs.httpproxy.Write(pkg.Sensor.id, pkg.Data); err != nil {
			cs.logger.Warnw(err.Error(), "sensor", pkg.Sensor.sid)
		}
	}

	// 数据包gRPC转发
	if cs.grpc != nil {
		if err := cs.grpc.Write(pkg.Sensor.id, pkg.Sensor.sid, pkg.Data); err != nil {
			cs.logger.Warnw(err.Error(), "sensor", pkg.Sensor.sid)
		}
	}

	/*
	 *	数据处理目前有3个:
	 *	1 - websocket 提供前段实时波动图展示
	 *	2 - file 文件存储，调试时开启，真正存储文件的服务是rawdb服务
	 *	3 - 时序数据库，温度、振动数据采样存储
	 */

	// 将数据送到websocket
	if cs.ws != nil {
		if err := cs.ws.Write(pkg.Sensor.id, frameBuff); err != nil {
			cs.logger.Errorw(err.Error(), "sensor", pkg.Sensor.sid)
		}
	}

	// 将数据送到文件
	if cs.file != nil {
		if err := cs.file.Write(pkg.Sensor.id, frameBuff); err != nil {
			cs.logger.Errorw(err.Error(), "sensor", pkg.Sensor.sid)
		}
	}

}

// ToHandle - 放入管道之前包处理
func (cs *Server) ToHandle(sensor *Sensor, data []byte, realtime int64) {
	// 负载均衡创建goroutine
	index := sensor.id & uint64(cs.config.GoroutineCount-1)
	v, loaded := cs.frameChans.Load(index)
	if !loaded {
		v = make(chan *Package, 1024*100)
		cs.frameChans.Store(index, v)
		cs.logger.Debugw("create package handle chan",
			"sensor", sensor.sid)

		go func() {
			handleFrameBuff := protocols.NewDefaultFrame()
			for p := range v.(chan *Package) {
				cs.handlePackage(handleFrameBuff, p)
			}
		}()
	}

	// 解包，对齐时间替换包号，置包号连续标志位
	/*
		_, err := cs.decodePackage(data, sensor, realtime)
		if err != nil {
			cs.logger.Errorw(err.Error())
			return
		}
	*/

	// 数据包入管道
	v.(chan *Package) <- &Package{
		Sensor:   sensor,
		Realtime: realtime,
		Data:     data,
	}

	cs.metrics.queuePkts.WithLabelValues(sensor.sid).Inc()
}

// BufferToHandle - 遍历缓存包处理, 这个是启动补包逻辑处理
func (cs *Server) BufferToHandle(sensor *Sensor, isResend bool) {
	var first int64
	var last int64

	for i := 0; i < len(sensor.buffArray); i++ {
		f := sensor.buffArray[i]

		// 判断包号中断
		if f.seq > sensor.lastSequence+1 && isResend {
			return
		}

		// 调试
		if i == 0 {
			first = f.seq
		}
		last = f.seq

		// 出队一个Frame包, 提交处理
		data := make([]byte, f.length)
		n, err := sensor.buff.Read(data)
		if err != nil {
			cs.logger.Errorw(err.Error())
		}
		if n != f.length {
			cs.logger.Errorw("read buff", "sensor", sensor.sid)
		}
		cs.ToHandle(sensor, data, f.realtime)
		sensor.buffArray = append(sensor.buffArray[:i], sensor.buffArray[i+1:]...)
		i--
	}

	if first > 0 {
		cs.logger.Debugw("resend buffer",
			"range", fmt.Sprintf("%d-%d", first, last),
			"left", len(sensor.buffArray),
			"sensor", sensor.sid)
	}
}

// ResendHandle - 重发判断
func (cs *Server) ResendHandle(seq int64, sensor *Sensor, data []byte) ([]byte, bool) {
	if seq > sensor.lastSequence+1 && sensor.lastSequence > 0 {

		// 判断缓存包是否到上线
		if len(sensor.buffArray) >= 10000 || sensor.redo > 5 {

			// 出队一个包, 提交处理
			buffdata := make([]byte, sensor.buffArray[0].length)
			n, err := sensor.buff.Read(buffdata)
			if err != nil {
				cs.logger.Errorw(err.Error())
			}
			if n != sensor.buffArray[0].length {
				cs.logger.Errorw("read buff", "sensor", sensor.sid)
			}
			sensor.buffArray = sensor.buffArray[1:]
			sensor.redo = 0

			cs.logger.Debugw("resend timeout",
				"seq", sensor.buffArray[0].seq,
				"length", sensor.buffArray[0].length,
				"sensor", sensor.sid)

			cs.ToHandle(sensor, buffdata, sensor.buffArray[0].realtime)

			// 处理缓存中包号连续的数据包
			cs.BufferToHandle(sensor, true)
		}

		// 重新判断包号是否断
		if seq > sensor.lastSequence+1 {
			// 缓存当前包，发送请求重发的包号
			_, err := sensor.buff.Write(data)
			if err != nil {
				cs.logger.Errorw(err.Error())
			}
			sensor.buffArray = append(sensor.buffArray, FrameBuff{
				seq:      seq,
				length:   len(data),
				realtime: sensor.lasttime,
			})

			if sensor.redo == 0 || time.Now().Unix()-sensor.reqtime > 5 {
				// 累计请求次数
				sensor.redo++

				cs.logger.Debugw("resend req",
					"req", sensor.lastSequence+1,
					"next", sensor.buffArray[0].seq,
					"wait", len(sensor.buffArray),
					"redo", sensor.redo,
					"sensor", sensor.sid)

				sensor.req = sensor.lastSequence + 1
				sensor.reqtime = time.Now().Unix()

				// 统计
				if sensor.redo == 1 {
					cs.metrics.resend.WithLabelValues(sensor.sid, "total").Inc()
				}

				// 发送请求
				return sensor.createFrame(uint64(sensor.lastSequence + 1)), true
			}
			return nil, true
		}
	}

	// 处理当前接收到的包之前，重置请求次数
	sensor.redo = 0

	// 如果当前包号，小于上一个包号，先处理缓存内所有的包
	if seq <= sensor.lastSequence {
		cs.BufferToHandle(sensor, false)
	}

	if seq == sensor.req {

		// 统计
		cs.metrics.resend.WithLabelValues(sensor.sid, "success").Inc()

		cs.logger.Debugw("resend ack",
			"request", sensor.lastSequence+1,
			"redo", sensor.redo,
			"sensor", sensor.sid)
	}
	return nil, false
}
