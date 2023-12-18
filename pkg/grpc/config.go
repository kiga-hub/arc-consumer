package grpc

import "github.com/spf13/viper"

const (
	// KeyGRPCEnable -
	KeyGRPCEnable = "grpc.enable"
	// KeyGRPCServer for data transfer
	KeyGRPCServer = "grpc.server"
)

var defaultConfig = Config{
	Enable: false,
	Server: "localhost:8080",
}

// Config struct grpc配置信息结构
type Config struct {
	Enable bool   `toml:"enable"`
	Server string `toml:"server"`
}

// SetDefaultConfig - 设置grpc配置参数
func SetDefaultConfig() {
	viper.SetDefault(KeyGRPCEnable, defaultConfig.Enable)
	viper.SetDefault(KeyGRPCServer, defaultConfig.Server)
}

// GetConfig - 获取grpc配置参数
// @return Config grpc配置数据结构
func GetConfig() *Config {
	return &Config{
		Enable: viper.GetBool(KeyGRPCEnable),
		Server: viper.GetString(KeyGRPCServer),
	}
}
