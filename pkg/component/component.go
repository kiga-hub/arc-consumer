package component

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	platformConf "github.com/kiga-hub/arc/conf"
	"github.com/kiga-hub/arc/configuration"
	"github.com/kiga-hub/arc/logging"
	logConf "github.com/kiga-hub/arc/logging/conf"
	"github.com/kiga-hub/arc/micro"
	microComponent "github.com/kiga-hub/arc/micro/component"
	"github.com/kiga-hub/arc/micro/conf"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/spf13/viper"

	"github.com/kiga-hub/arc-consumer/pkg/api"
	"github.com/kiga-hub/arc-consumer/pkg/file"
	"github.com/kiga-hub/arc-consumer/pkg/gnet"
	"github.com/kiga-hub/arc-consumer/pkg/goss"
	"github.com/kiga-hub/arc-consumer/pkg/grpc"
	"github.com/kiga-hub/arc-consumer/pkg/httpproxy"
	"github.com/kiga-hub/arc-consumer/pkg/proxy"
	"github.com/kiga-hub/arc-consumer/pkg/ws"
)

// ArcConsumerElementKey is Element Key for arc-consumer
var ArcConsumerElementKey = micro.ElementKey("ArcConsumerComponent")

// ArcConsumerComponent is Component for TaskMgmt
type ArcConsumerComponent struct {
	micro.EmptyComponent
	cluster       string
	privateIP     string
	config        *conf.BasicConfig
	logger        logging.ILogger
	nacosClient   *configuration.NacosClient
	gossipKVCache *microComponent.GossipKVCacheComponent
	ws            ws.Handler
	gnet          gnet.Handler
	grpc          grpc.Handler
	api           api.Handler
	kvCache       goss.Handler
}

// Name of the component
func (c *ArcConsumerComponent) Name() string {
	return "ArcConsumerComponent"
}

// PreInit called before Init()
func (c *ArcConsumerComponent) PreInit(ctx context.Context) error {
	file.SetDefaultConfig()
	gnet.SetDefaultConfig()
	grpc.SetDefaultConfig()
	proxy.SetDefaultConfig()
	httpproxy.SetDefaultConfig()
	ws.SetDefaultConfig()
	return nil
}

// Init the component
func (c *ArcConsumerComponent) Init(server *micro.Server) (err error) {
	c.cluster = server.PrivateCluster
	c.privateIP = server.PrivateIP.String()
	c.config = conf.GetBasicConfig()
	spew.Dump(c.config) // 打印基础配置信息

	// 动态配置接口
	elNacos := server.GetElement(&micro.NacosClientElementKey)
	if elNacos != nil {
		c.nacosClient = elNacos.(*configuration.NacosClient)
	}

	// 获取日志接口
	elLogger := server.GetElement(&micro.LoggingElementKey)
	if elLogger != nil {
		c.logger = elLogger.(logging.ILogger)
		spew.Dump(logConf.GetLogConfig()) // 打印日志配置信息
	}

	// 集群一致性键值接口
	elkvcache := server.GetElement(&micro.GossipKVCacheElementKey)
	if elkvcache != nil {
		c.gossipKVCache = elkvcache.(*microComponent.GossipKVCacheComponent)
		// 键值上报服务初始化
		c.kvCache = goss.New(
			goss.WithKVCache(c.gossipKVCache),
			goss.WithLogger(c.logger),
		)
	}

	// 初始化websocket服务
	if c.ws, err = ws.New(
		ws.WithLogger(c.logger),
	); err != nil {
		return err
	}

	// 初始化grpck客户端服务，目前用于转发数据到rawdb
	c.grpc = grpc.New(grpc.WithLogger(c.logger))

	// 初始化tcp服务
	if c.gnet, err = gnet.New(
		gnet.WithLogger(c.logger),
		gnet.WithGrpc(c.grpc),
		gnet.WithWebSocket(c.ws),
		gnet.WithKVCache(c.kvCache),
	); err != nil {
		return err
	}

	// 初始化web api接口服务
	c.api = api.New(
		api.WithLogger(c.logger),
		api.WithWebSocket(c.ws),
		api.WithGossipKVCache(c.gossipKVCache),
		api.WithGnet(c.gnet),
	)

	return nil
}

// SetDynamicConfig 加载nacos动态配置回调
func (c *ArcConsumerComponent) SetDynamicConfig(nf *platformConf.NodeConfig) error {
	if nf == nil || nf.DataTransfer == nil {
		return nil
	}

	// 开启接收服务
	if nf.DataTransfer.EnableReceive {
		if nf.DataTransfer.ReceiveConfig != nil {
			viper.Set(gnet.KeyEnableCalculateTimestamp, nf.DataTransfer.ReceiveConfig.EnableCalculateTimestamp) // 时间对齐配置
			viper.Set(gnet.KeyEnableCRCCheck, nf.DataTransfer.ReceiveConfig.EnableCRCCheck)                     // crc 校验配置
		}
	}

	// 开启数据管理服务
	if nf.DataTransfer.EnableReadWrite {
		viper.Set(grpc.KeyGRPCEnable, true) // grpc 将数据发送给rawdb
	} else {
		viper.Set(grpc.KeyGRPCEnable, false)
	}

	// 开启转发服务
	if nf.DataTransfer.EnableSend {
		viper.Set(proxy.KeyProxyEnable, true)
		if nf.DataTransfer.SendConfig != nil && nf.DataTransfer.SendConfig.RemoteReceiverAddr != "" {
			viper.Set(proxy.KeyProxyParent, nf.DataTransfer.SendConfig.RemoteReceiverAddr) // 远端服务地址
		}
	} else {
		viper.Set(proxy.KeyProxyEnable, false)
	}
	return nil
}

// OnConfigChanged 动态配置nacos修改回调函数
func (c *ArcConsumerComponent) OnConfigChanged(*platformConf.NodeConfig) error {
	return micro.ErrNeedRestart
}

// SetupHandler 安装路由
func (c *ArcConsumerComponent) SetupHandler(root echoswagger.ApiRoot, base string) error {
	root.Echo().Static(c.config.APIRoot, "./web") // 波动图静态页面
	c.api.Setup(root, base)                       // restful接口
	return nil
}

// Start the component
func (c *ArcConsumerComponent) Start(ctx context.Context) error {
	// 集群定时上报模块启动
	if c.kvCache != nil {
		go c.kvCache.Start(ctx)
	}

	// websocket模块启动
	if c.ws != nil {
		go c.ws.Start(ctx)
	}

	// 数据接收模块启动
	go func() {
		if err := c.gnet.Start(ctx); err != nil {
			panic(err)
		}
	}()

	// 判断服务是否在集群内，不在集群内直接连接
	if c.gossipKVCache == nil {
		// 不在集群内直接连接远程grpc服务
		if c.grpc != nil {
			go c.grpc.Start(ctx)
			c.grpc.ReConnect()
		}
		return nil
	}

	// 回调事件
	c.setEventHook()
	return nil
}

// Stop the component
func (c *ArcConsumerComponent) Stop(ctx context.Context) error {
	// 停止数据接收模块
	if err := c.gnet.Stop(); err != nil {
		c.logger.Errorw("stop gnet", "err", err)
	}

	// 停止grpc服务
	if c.grpc != nil {
		c.grpc.Stop()
	}
	return nil
}
