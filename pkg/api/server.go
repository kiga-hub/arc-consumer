package api

import (
	"github.com/kiga-hub/arc-consumer/pkg/gnet"
	"github.com/kiga-hub/arc-consumer/pkg/ws"
	"github.com/kiga-hub/arc/logging"
	microComponent "github.com/kiga-hub/arc/micro/component"
	"github.com/kiga-hub/arc/micro/conf"
	"github.com/pangpanglabs/echoswagger/v2"
)

// Handler -
type Handler interface {
	Setup(echoswagger.ApiRoot, string)
}

// Server - api处理器
type Server struct {
	logger          logging.ILogger
	ws              ws.Handler
	gnet            gnet.Handler
	gossipKVCache   *microComponent.GossipKVCacheComponent
	selfServiceName string
}

// New - 初始化
// @param opts Option 设置选项的函数，可变参数
// @return Handler api处理器结构
func New(opts ...Option) Handler {
	srv := loadOptions(opts...)
	srv.selfServiceName = conf.GetBasicConfig().Service
	return srv
}
