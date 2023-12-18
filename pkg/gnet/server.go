package gnet

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/protocols"
	"github.com/panjf2000/gnet"

	"github.com/kiga-hub/arc-consumer/pkg/file"
	"github.com/kiga-hub/arc-consumer/pkg/goss"
	"github.com/kiga-hub/arc-consumer/pkg/grpc"
	"github.com/kiga-hub/arc-consumer/pkg/httpproxy"
	"github.com/kiga-hub/arc-consumer/pkg/proxy"
	"github.com/kiga-hub/arc-consumer/pkg/ws"
)

// Handler - gnet 接口
type Handler interface {
	Start(context.Context) error
	Stop() error
	GetSensors() []uint64
	GetSensor(id uint64) *Sensor
}

// Server - 服务结构
type Server struct {
	*gnet.EventServer
	sensors    *sync.Map
	frameChans *sync.Map
	tmap       *sync.Map
	config     *Config
	logger     logging.ILogger
	ws         ws.Handler
	grpc       grpc.Handler
	file       file.Handler
	kvCache    goss.Handler
	proxy      proxy.Handler
	httpproxy  httpproxy.Handler
	metrics    *Metrics
}

// New  - 初始化结构
func New(opts ...Option) (Handler, error) {
	srv := loadOptions(opts...)
	var err error

	spew.Dump(srv.config)

	srv.tmap = new(sync.Map)

	// 统计信息上报
	srv.metrics = NewMetrics()

	// 检查goroutine count
	if srv.config.GoroutineCount <= 0 ||
		(srv.config.GoroutineCount&(srv.config.GoroutineCount-1)) != 0 {
		return nil, fmt.Errorf("config goroutine count err")
	}

	if srv.grpc != nil {
		srv.grpc.SetMask(uint64(srv.config.GoroutineCount - 1))
	}

	// init file
	if srv.file, err = file.New(
		file.WithLogger(srv.logger),
	); err != nil {
		return nil, err
	}

	// init proxy
	if srv.proxy, err = proxy.New(uint64(srv.config.GoroutineCount-1), proxy.WithLogger(srv.logger)); err != nil {
		return nil, err
	}

	// init proxy
	if srv.httpproxy, err = httpproxy.New(httpproxy.WithLogger(srv.logger)); err != nil {
		return nil, err
	}
	return srv, nil
}

// GetSensors - 获取设备id列表
func (cs *Server) GetSensors() []uint64 {
	var res []uint64
	cs.sensors.Range(func(key, value interface{}) bool {
		res = append(res, value.(*Sensor).id)
		return true
	})
	return res
}

// Start - 启动服务
func (cs *Server) Start(ctx context.Context) error {
	codec := &protocols.Coder{}
	codec.IsCrcCheck = cs.config.EnableCRCCheck
	if cs.file != nil {
		go cs.file.Start(ctx)
	}

	// proxy start
	if cs.proxy != nil {
		go cs.proxy.Start(ctx)
	}

	// httpproxy start
	if cs.httpproxy != nil {
		go cs.httpproxy.Start(ctx)
	}

	addr := fmt.Sprintf("%s://:%d", cs.config.NetType, cs.config.Port)
	err := gnet.Serve(cs, addr,
		// gnet.WithLoadBalancing(gnet.SourceAddrHash),
		gnet.WithLogger(cs.logger),
		gnet.WithMulticore(true),
		gnet.WithNumEventLoop(runtime.GOMAXPROCS(0)/2+1),
		gnet.WithTCPKeepAlive(time.Millisecond*time.Duration(cs.config.Keepalive)),
		gnet.WithCodec(codec))
	if err != nil {
		cs.logger.Errorw(err.Error(), "addr", addr)
	}
	return err
}

// Stop - 停止服务
func (cs *Server) Stop() error {
	if cs.file != nil {
		cs.file.Stop()
	}
	// proxy stop
	if cs.proxy != nil {
		cs.proxy.Stop()
	}
	// httpproxy stop
	if cs.httpproxy != nil {
		cs.httpproxy.Stop()
	}
	addr := fmt.Sprintf("%s://:%d", cs.config.NetType, cs.config.Port)
	return gnet.Stop(context.Background(), addr)
}

// GetSensor - 获取一个传感器结构信息
func (cs *Server) GetSensor(id uint64) *Sensor {
	v, ok := cs.sensors.Load(id)
	if !ok {
		return nil
	}
	return v.(*Sensor)
}
