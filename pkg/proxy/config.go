package proxy

import "github.com/spf13/viper"

const (
	// KeyProxyEnable -
	KeyProxyEnable = "proxy.enable"
	// KeyProxyParent -
	KeyProxyParent = "proxy.parent"
	proxytimeout   = "proxy.timeout"
	reconnInterval = "proxy.interval"
	nettype        = "proxy.net_type"
)

var defaultConfig = Config{
	Enable:   false,
	Parent:   "localhost:8972",
	Timeout:  1000,
	NetType:  "tcp",
	Interval: 1000,
}

// Config - 代理配置信息结构
type Config struct {
	Enable   bool   `toml:"enable"`
	Parent   string `toml:"parent"`
	Timeout  int    `toml:"timeout"`
	NetType  string `toml:"net_type" json:"net_type,omitempty"`
	Interval int    `toml:"interval"`
}

// SetDefaultConfig - 设置代理配置
func SetDefaultConfig() {
	viper.SetDefault(KeyProxyEnable, defaultConfig.Enable)
	viper.SetDefault(KeyProxyParent, defaultConfig.Parent)
	viper.SetDefault(proxytimeout, defaultConfig.Timeout)
	viper.SetDefault(reconnInterval, defaultConfig.Interval)
	viper.SetDefault(nettype, defaultConfig.NetType)
}

// GetConfig - 获取代理配置
// @return *Config 代理配置信息结构
func GetConfig() *Config {
	return &Config{
		NetType:  viper.GetString(nettype),
		Enable:   viper.GetBool(KeyProxyEnable),
		Parent:   viper.GetString(KeyProxyParent),
		Timeout:  viper.GetInt(proxytimeout),
		Interval: viper.GetInt(reconnInterval),
	}
}
