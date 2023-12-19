package simulate

import (
	"context"
	"fmt"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/protocols"

	"github.com/kiga-hub/arc-consumer/pkg/goss"
	"github.com/kiga-hub/arc-consumer/pkg/grpc"
)

// Sensor - 传感器结构
type Sensor struct {
	id  uint64 // 编号
	sid string // 字符串编号

}

// Handler - simulate 接口
type Handler interface {
	Start(context.Context) error
	Stop() error
}

// Server - 服务结构
type Server struct {
	// *simulate.EventServer
	sensors    *sync.Map
	frameChans *sync.Map
	tmap       *sync.Map
	config     *Config
	logger     logging.ILogger
	grpc       grpc.Handler
	kvCache    goss.Handler
}

// New  - 初始化结构
func New(opts ...Option) (Handler, error) {
	srv := loadOptions(opts...)

	spew.Dump(srv.config)

	srv.tmap = new(sync.Map)

	// 检查goroutine count
	if srv.config.GoroutineCount <= 0 ||
		(srv.config.GoroutineCount&(srv.config.GoroutineCount-1)) != 0 {
		return nil, fmt.Errorf("config goroutine count err")
	}

	if srv.grpc != nil {
		srv.grpc.SetMask(uint64(srv.config.GoroutineCount - 1))
	}

	return srv, nil
}

// Start - 启动服务
func (cs *Server) Start(ctx context.Context) error {
	codec := &protocols.Coder{}
	codec.IsCrcCheck = cs.config.EnableCRCCheck

	go func() {
		cs.Producer()
	}()

	return nil
}

// Stop - 停止服务
func (cs *Server) Stop() error {
	return nil
}

func (cs *Server) Producer() {
	sensor := &Sensor{
		id:  1,
		sid: "94C96000C248",
	}
	data := make([]byte, 1024)
	cs.ToHandle(sensor, data)
}
