package simulate

import (
	"fmt"

	"github.com/kiga-hub/arc/protocols"
)

// Package - 处理包结构
type Package struct {
	Sensor *Sensor // 传感器
	Data   []byte  // 数据
}

// decodePackage -
func (cs *Server) decodePackage(frameBuff *protocols.Frame, data []byte, sensor *Sensor) (int64, error) {
	// 从Frame包获取数据段(获取Frame包时间是根据时间计算的)
	var err error
	var sa *protocols.SegmentArc
	for _, stype := range frameBuff.DataGroup.STypes {
		switch stype {
		case protocols.STypeArc:
			sa, err = frameBuff.DataGroup.GetArcSegment()
			if err != nil {
				return 0, err
			}
		}
	}

	// Frame包内必须有数据段
	if sa == nil {
		return 0, fmt.Errorf("%012X package not find arc data", sensor.id)
	}

	// 根据数据，计算数据时间（微秒）
	dataSize := int64(len(sa.Data))
	return dataSize, nil
}

// 从管道获取package结构，包处理
func (cs *Server) handlePackage(frameBuff *protocols.Frame, pkg *Package) {
	// 集群同步上报传感器编号
	if cs.kvCache != nil {
		if err := cs.kvCache.Sync(pkg.Sensor.id); err != nil {
			cs.logger.Errorw(err.Error(), "sensor", pkg.Sensor.sid)
		}
	}

	// 解包
	if err := frameBuff.Decode(pkg.Data); err != nil {
		cs.logger.Errorw(err.Error(), "sensor", pkg.Sensor.sid)
		return
	}

	// 时间对齐，统计检查
	_, err := cs.decodePackage(frameBuff, pkg.Data, pkg.Sensor)
	if err != nil {
		cs.logger.Errorw(err.Error())
		return
	}

	cs.tmap.Store(pkg.Sensor.id, frameBuff.Timestamp)

	// 数据包gRPC转发
	if cs.grpc != nil {
		if err := cs.grpc.Write(pkg.Sensor.id, pkg.Sensor.sid, pkg.Data); err != nil {
			cs.logger.Warnw(err.Error(), "sensor", pkg.Sensor.sid)
		}
	}
}

// ToHandle - 放入管道之前包处理
func (cs *Server) ToHandle(sensor *Sensor, data []byte) {
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

	// 数据包入管道
	v.(chan *Package) <- &Package{
		Sensor: sensor,
		Data:   data,
	}
}
