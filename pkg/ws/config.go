package ws

import "github.com/spf13/viper"

const (
	wsenable     = "ws.enable"
	wsfill       = "ws.fill"
	wsfillenable = "ws.fill_enable"
)

var defaultConfig = Config{
	Enable:     true,
	FillEnable: true,
	Fill:       10,
}

// Config struct
type Config struct {
	Enable     bool `toml:"enable"`
	FillEnable bool `toml:"fill_enable"`
	Fill       int  `toml:"fill"`
}

// SetDefaultConfig - 默认配置
func SetDefaultConfig() {
	viper.SetDefault(wsenable, defaultConfig.Enable)
	viper.SetDefault(wsfillenable, defaultConfig.FillEnable)
	viper.SetDefault(wsfill, defaultConfig.Fill)
}

// GetConfig - 获取配置值
func GetConfig() *Config {
	return &Config{
		Enable:     viper.GetBool(wsenable),
		FillEnable: viper.GetBool(wsfillenable),
		Fill:       viper.GetInt(wsfill),
	}
}
