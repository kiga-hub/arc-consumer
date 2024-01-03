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
	"github.com/kiga-hub/arc-consumer/pkg/goss"
	"github.com/kiga-hub/arc-consumer/pkg/grpc"
	"github.com/kiga-hub/arc-consumer/pkg/simulate"
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
	simulate      simulate.Handler
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
	simulate.SetDefaultConfig()
	grpc.SetDefaultConfig()
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

	// 初始化grpck客户端服务，目前用于转发数据到arc-storage
	c.grpc = grpc.New(grpc.WithLogger(c.logger))

	// 初始化tcp服务
	if c.simulate, err = simulate.New(
		simulate.WithLogger(c.logger),
		simulate.WithGrpc(c.grpc),
		simulate.WithKVCache(c.kvCache),
	); err != nil {
		return err
	}

	// 初始化web api接口服务
	c.api = api.New(
		api.WithLogger(c.logger),
		api.WithGossipKVCache(c.gossipKVCache),
		api.WithSimulate(c.simulate),
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
			viper.Set(simulate.KeyEnableCRCCheck, nf.DataTransfer.ReceiveConfig.EnableCRCCheck) // crc 校验配置
		}
	}

	// 开启数据管理服务
	if nf.DataTransfer.EnableReadWrite {
		viper.Set(grpc.KeyGRPCEnable, true) // grpc 将数据发送给arc-storage
	} else {
		viper.Set(grpc.KeyGRPCEnable, false)
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

	// 数据接收模块启动
	go func() {
		if err := c.simulate.Start(ctx); err != nil {
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
	if err := c.simulate.Stop(); err != nil {
		c.logger.Errorw("stop simulate", "err", err)
	}

	// 停止grpc服务
	if c.grpc != nil {
		c.grpc.Stop()
	}
	return nil
}
