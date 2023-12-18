package api

import (
	"net/http"

	"github.com/pangpanglabs/echoswagger/v2"
)

// BaseResponse is the response
type BaseResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Setup - 接口服务设置
// @param root echoswagger.ApiRoot API接口
// @param base string 路由前缀
func (s *Server) Setup(root echoswagger.ApiRoot, base string) {
	if s.ws != nil {
		s.setupWebsocket(root, base) // websocket 接口
	}
}

// setupProfile - WebSocket服务
// @param root echoswagger.ApiRoot API接口
// @param base string 路由前缀
func (s *Server) setupWebsocket(root echoswagger.ApiRoot, base string) {
	g := root.Group("Websocket", base+"/ws")

	handlerFunc := s.collectorWs
	if s.gossipKVCache != nil {
		handlerFunc = s.gossipKVCache.SensorIDsHandlerWrapper(s.selfServiceName, s.collectorWs, s.handleWs)
	}

	g.GET("/collector", handlerFunc).
		AddParamQuery("", "collectorids", "collectorids", true).
		SetOperationId(`collector ws`).
		SetSummary("采集设备websocket").
		SetDescription(`获取采集设备websocket
		连接websocket成功后
		发送json数据格式：
		{
			"collectorids": "A00000000000" 	- 多个采集设备ID用逗号隔开
			"interval": 100						- 数据间隔（毫秒）
			"temperature": true					- 是否获取温度数据
			"vibrate": true						- 是否获取震动数据
			"vibrate_vatd": true				- 是否获取振动强度数据（需要配置启动引擎）
			"audio": true						- 是否获取音频数据
			"audio_vatd": true					- 是否获取音频强度数据（需要配置启动引擎）
			"audio_spark": true					- 是否获取音频电火花数据（需要配置启动引擎）
			"fill": true					    - 是否填充数据
		}
		`).
		AddResponse(http.StatusOK, `
		{
			"audio_engine":            - 音频过算法引擎后数据 
			{
				collectorid: - 采集设备ID
				"data":[{
					"time": 1615178437100,  		   	- 数据保存时间(毫秒时间戳)
					"db": 100.18087						- 音频强度
					"ind_impulse": 4.6229773, 		 	- 脉冲指标
					"ind_kurt": 3.4695735,				- 峭度指标
					"ind_margin": 7.070554,				- 裕度指标
					"ind_peak": 3.01643, 				- 峰值指标
					"ind_waveform": 1.532599,			- 波形指标
					"kurt": 0.74414146,					- 峭度
					"maxv": 2.0527596,					- 最大值
					"mean": -0.05275969,				- 均值
					"minv": -1.9472404,					- 最小值
					"peak": 2.0527596,					- 峰值
					"ppv": 4,							- 峰峰值
					"raw": 1,							- 原始采样值
					"rms": 36.8288346877526,			- 均方根（强度）
					"skew": -0.020787					- 偏度
					"std": 0.67847794,					- 标准差
					"varance": 0.4603323,				- 方差
					"spark": 1.8288346877526,			- 电火花事件识别值
				}],
			},
			"vibrate_engine":            - 振动过算法引擎后数据 
			{
				collectorid: - 采集设备ID
				"data":[{
					"time": 1615178437100,  		   	- 数据保存时间(毫秒时间戳)
					"acc": 0								- 加速度
					"ind_impulse": 4.6229773, 		 	- 脉冲指标
					"ind_kurt": 3.4695735,				- 峭度指标
					"ind_margin": 7.070554,				- 裕度指标
					"ind_peak": 3.01643, 				- 峰值指标
					"ind_waveform": 1.532599,			- 波形指标
					"kurt": 0.74414146,					- 峭度
					"maxv": 2.0527596,					- 最大值
					"mean": -0.05275969,				- 均值
					"minv": -1.9472404,					- 最小值
					"peak": 2.0527596,					- 峰值
					"ppv": 4,							- 峰峰值
					"raw": 1,							- 原始采样值
					"rms": 36.8288346877526,			- 均方根（强度）
					"skew": -0.020787					- 偏度
					"std": 0.67847794,					- 标准差
					"varance": 0.4603323,				- 方差
				},{
					"time": 1615178437100,  		   	- 数据保存时间(毫秒时间戳)
					"acc": 	0							- 加速度
					"ind_impulse": 4.6229773, 		 	- 脉冲指标
					"ind_kurt": 3.4695735,				- 峭度指标
					"ind_margin": 7.070554,				- 裕度指标
					"ind_peak": 3.01643, 				- 峰值指标
					"ind_waveform": 1.532599,			- 波形指标
					"kurt": 0.74414146,					- 峭度
					"maxv": 2.0527596,					- 最大值
					"mean": -0.05275969,				- 均值
					"minv": -1.9472404,					- 最小值
					"peak": 2.0527596,					- 峰值
					"ppv": 4,							- 峰峰值
					"raw": 1,							- 原始采样值
					"rms": 36.8288346877526,			- 均方根（强度）
					"skew": -0.020787					- 偏度
					"std": 0.67847794,					- 标准差
					"varance": 0.4603323,				- 方差
				},{
					"time": 1615178437100,  		   	- 数据保存时间(毫秒时间戳)
					"acc": 0							- 加速度
					"ind_impulse": 4.6229773, 		 	- 脉冲指标
					"ind_kurt": 3.4695735,				- 峭度指标
					"ind_margin": 7.070554,				- 裕度指标
					"ind_peak": 3.01643, 				- 峰值指标
					"ind_waveform": 1.532599,			- 波形指标
					"kurt": 0.74414146,					- 峭度
					"maxv": 2.0527596,					- 最大值
					"mean": -0.05275969,				- 均值
					"minv": -1.9472404,					- 最小值
					"peak": 2.0527596,					- 峰值
					"ppv": 4,							- 峰峰值
					"raw": 1,							- 原始采样值
					"rms": 36.8288346877526,			- 均方根（强度）
					"skew": -0.020787					- 偏度
					"std": 0.67847794,					- 标准差
					"varance": 0.4603323,				- 方差
				}],
			},
			"vibrate":            	- 震动原始数据
			{
				"collectorid": "A00000000000"
				"samplerate": 3200
				"range": 16
				"resolution": 13
				"offset_x": 0
				"offset_y": 0
				"offset_z": 0,
				"data": [{
					"time": 1615178437100,
				}],
			},
			"temperature": {
				collectorid: "A00000000000",
				"data":[{
					"time": 1615178437100,
					"t":25.375
				}],
			},
			"audio": {
				"collectorid": "A00000000000",
				"samplerate": 32000,
				"bits": 16,
				"channel": 1,
				"data":[{
					"time": 1615178437100,
				}]
			}
		}
		`, nil, nil)

	g.GET("/sensorids", s.collectorWssensorids).
		SetOperationId(`collector ws sensorids`).
		SetSummary("采集设备websocket请求的传感器ID数组").
		SetDescription(`获取采集设备websocket传感器数组`).
		AddResponse(http.StatusOK, ` `, nil, nil)

	g.GET("/upload", s.collectorDataUpload).
		SetOperationId(`collector upload data`).
		SetSummary("上传传感器数据").
		SetDescription(`http上传传感器多个Frame数据`).
		AddResponse(http.StatusOK, ` `, nil, nil)
}
