package simulate

import "github.com/spf13/viper"

const (
	nettype           = "service.net_type"
	host              = "service.host"
	devicehost        = "service.device_host"
	port              = "service.port"
	keepalive         = "service.keepalive"
	goroutineCount    = "service.goroutine_count"
	KeyEnableCRCCheck = "service.enable_crc_check"
	proxyTimealign    = "service.proxy_timealign"
)

// 配置默认值 - 最低优先级
var defaultConfig = Config{
	NetType:        "tcp",
	Host:           "127.0.0.1",
	DeviceHost:     "127.0.0.1",
	Port:           8972,
	Keepalive:      60000,
	GoroutineCount: 8,
	EnableCRCCheck: true,
	ProxyTimealign: true,
}

// Config - 配置结构
type Config struct {
	NetType        string `toml:"net_type" json:"net_type,omitempty"`
	Host           string `toml:"host" json:"host,omitempty"`
	DeviceHost     string `toml:"device_host" json:"device_host,omitempty"`
	Port           int    `toml:"port" json:"port,omitempty"`
	Keepalive      int    `toml:"keepalive" json:"keepalive,omitempty"`
	GoroutineCount int    `toml:"goroutine_count" json:"goroutine_count,omitempty"`
	EnableCRCCheck bool   `toml:"enable_crc_check" json:"enable_crc_check,omitempty"`
	ProxyTimealign bool   `toml:"proxy_timealign" json:"proxy_timealign"`
}

// SetDefaultConfig - 设置默认配置
func SetDefaultConfig() {
	viper.SetDefault(nettype, defaultConfig.NetType)
	viper.SetDefault(host, defaultConfig.Host)
	viper.SetDefault(devicehost, defaultConfig.DeviceHost)
	viper.SetDefault(port, defaultConfig.Port)
	viper.SetDefault(keepalive, defaultConfig.Keepalive)
	viper.SetDefault(goroutineCount, defaultConfig.GoroutineCount)
	viper.SetDefault(KeyEnableCRCCheck, defaultConfig.EnableCRCCheck)
	viper.SetDefault(proxyTimealign, defaultConfig.ProxyTimealign)
}

// GetConfig - 获取当前配置
func GetConfig() *Config {
	return &Config{
		NetType:        viper.GetString(nettype),
		Host:           viper.GetString(host),
		DeviceHost:     viper.GetString(devicehost),
		Port:           viper.GetInt(port),
		Keepalive:      viper.GetInt(keepalive),
		GoroutineCount: viper.GetInt(goroutineCount),
		EnableCRCCheck: viper.GetBool(KeyEnableCRCCheck),
		ProxyTimealign: viper.GetBool(proxyTimealign),
	}
}
