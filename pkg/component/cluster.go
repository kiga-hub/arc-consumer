package component

import (
	"fmt"
	"strings"

	"github.com/kiga-hub/arc-consumer/pkg/grpc"
	microComponent "github.com/kiga-hub/arc/micro/component"
	"github.com/spf13/viper"
)

const rawdbServiceName = "arc-consumer"

// findRawDB 从集群获取，rawdb地址
func (c *ArcConsumerComponent) findRawDB(addr string) (string, error) {
	if c.gossipKVCache == nil {
		return strings.Split(addr, ":")[0], nil
	}
	_, pip, err := c.gossipKVCache.FindMemberIPs(c.cluster, rawdbServiceName)
	return pip, err
}

func (c *ArcConsumerComponent) setEventHook() {
	grpcConfig := grpc.GetConfig()

	// 本服务启动，加入集群后，回调函数
	c.gossipKVCache.OnJoinCluster = func() {
		if c.grpc != nil {
			// 获取rawdb地址
			ip, err := c.findRawDB(grpcConfig.Server)
			if err != nil {
				c.logger.Errorf("not find raw db service")
			} else {
				// 更新rawdb配置
				addr := fmt.Sprintf("%s:8080", ip)
				viper.Set(grpc.KeyGRPCServer, addr)

				// 连接rawdb
				c.grpc.ReConnect()
				c.logger.Infof("set rawdb address to %s for %s", addr, c.privateIP)
			}
		}
	}

	// 集群中有掉线的服务回调
	c.gossipKVCache.OnNodeLeave = func(n *microComponent.GossipKVCacheNodeMeta) {
		// 判断集群掉线是不是rawdb服务
		if n.PrivateCluster != c.cluster || n.ServiceName != rawdbServiceName {
			return
		}
		c.logger.Infof("on node leave %s, %s", n.PrivateCluster, n.ServiceName)
		// 断开rawdb连接
		if grpcConfig.Enable {
			c.grpc.Disconnect()
		}
	}

	// 集群中有上线的服务回调
	c.gossipKVCache.OnNodeJoin = func(n *microComponent.GossipKVCacheNodeMeta) {
		// 判断集群上线是不是rawdb服务
		if n.PrivateCluster != c.cluster || n.ServiceName != rawdbServiceName {
			return
		}
		c.logger.Infof("on node join %s, %s", n.PrivateCluster, n.ServiceName)

		// 连接rawdb
		if grpcConfig.Enable {
			addr := fmt.Sprintf("%s:8080", n.PrivateIP)
			viper.Set(grpc.KeyGRPCServer, addr)
			grpcConfig.Server = addr
			c.grpc.ReConnect()
			c.logger.Infof("set rawdb address to %s for %s", addr, c.privateIP)
		}
	}
}
