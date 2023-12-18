package httpproxy

import "github.com/spf13/viper"

const (
	// HTTPProxyEnable -
	HTTPProxyEnable = "httpproxy.enable"
	// HTTPProxyParent -
	HTTPProxyParent = "httpproxy.parent"
	httptimeout     = "httpproxy.timeout"
	httpsize        = "httpproxy.size"
)

var defaultConfig = Config{
	Enable:  false,
	Parent:  "http://localhost/api/data/v1/realtime",
	Timeout: 100,
	Size:    10240,
}

// Config - 代理配置信息结构
type Config struct {
	Enable  bool   `toml:"enable"`
	Parent  string `toml:"parent"`
	Timeout int    `toml:"timeout"`
	Size    int    `toml:"size"`
}

// SetDefaultConfig - 设置代理配置
func SetDefaultConfig() {
	viper.SetDefault(HTTPProxyEnable, defaultConfig.Enable)
	viper.SetDefault(HTTPProxyParent, defaultConfig.Parent)
	viper.SetDefault(httptimeout, defaultConfig.Timeout)
	viper.SetDefault(httpsize, defaultConfig.Size)
}

// GetConfig - 获取代理配置
// @return *Config 代理配置信息结构
func GetConfig() *Config {
	return &Config{
		Enable:  viper.GetBool(HTTPProxyEnable),
		Parent:  viper.GetString(HTTPProxyParent),
		Timeout: viper.GetInt(httptimeout),
		Size:    viper.GetInt(httpsize),
	}
}
