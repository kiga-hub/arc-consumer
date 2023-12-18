package file

import (
	"github.com/spf13/viper"
)

const (
	fileAEnable     = "file.a_enable"
	fileNEnable     = "file.n_enable"
	fileVEnable     = "file.v_enable"
	fileTEnable     = "file.t_enable"
	fileMVEnable    = "file.mv_enable"
	filedirent      = "file.dirent"
	filesubdirent   = "file.subdirent"
	filetimeout     = "file.timeout"
	fileduratoinmin = "file.duration_min"
)

var defaultConfig = Config{
	AEnable:     false,
	NEnable:     false,
	VEnable:     false,
	TEnable:     false,
	MVEnable:    false,
	DurationMin: 1,
	Dirent:      "./data",
	SubDirent:   "day", // year/month/day/none
	Timeout:     10000,
}

// Config struct
type Config struct {
	AEnable     bool   `toml:"a_enable"`
	NEnable     bool   `toml:"n_enable"`
	VEnable     bool   `toml:"v_enable"`
	TEnable     bool   `toml:"t_enable"`
	MVEnable    bool   `toml:"mv_enable"`
	DurationMin int    `toml:"duration_min"`
	Dirent      string `toml:"dirent"`
	SubDirent   string `toml:"subdirent"`
	Timeout     int    `toml:"timeout"`
}

// SetDefaultConfig - 设置文件存储配置
func SetDefaultConfig() {
	viper.SetDefault(fileAEnable, defaultConfig.AEnable)
	viper.SetDefault(fileNEnable, defaultConfig.NEnable)
	viper.SetDefault(fileVEnable, defaultConfig.VEnable)
	viper.SetDefault(fileTEnable, defaultConfig.TEnable)
	viper.SetDefault(fileMVEnable, defaultConfig.MVEnable)
	viper.SetDefault(filedirent, defaultConfig.Dirent)
	viper.SetDefault(filesubdirent, defaultConfig.SubDirent)
	viper.SetDefault(fileduratoinmin, defaultConfig.DurationMin)
	viper.SetDefault(filetimeout, defaultConfig.Timeout)
}

// GetConfig - 获取文件存储配置
// @return Config file配置数据结构
func GetConfig() *Config {
	return &Config{
		AEnable:     viper.GetBool(fileAEnable),
		NEnable:     viper.GetBool(fileNEnable),
		VEnable:     viper.GetBool(fileVEnable),
		TEnable:     viper.GetBool(fileTEnable),
		MVEnable:    viper.GetBool(fileMVEnable),
		DurationMin: viper.GetInt(fileduratoinmin),
		Timeout:     viper.GetInt(filetimeout),
		Dirent:      viper.GetString(filedirent),
		SubDirent:   viper.GetString(filesubdirent),
	}
}
