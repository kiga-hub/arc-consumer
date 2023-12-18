# 数据接收

| 版本 | 修改内容                                   | 创建人           | 创建时间   |
| ---- | ------------------------------------------ | ---------------- | ---------- |
| V0.1 | 第一版                                     | 于超           | 2022-02-22 |

## 目的

从数据接收端获取数据接口,websocket实时数据接口,旧的缓存接口形式已经废弃

## 术语和定义

## 接口系统的能力范围

## 主要流程

- 前段页面访问websocket波动图

## 接口整体设计

通讯协议采用 HTTP 1.1，服务的根路径为`/api/data/v1/realtime`

### 公共请求消息头

下表列出了所有声纹数据库接口所携带的公共头域。HTTP 协议的标准头域不再这里列出。

| 消息头（Header） | 是否必须 | 说明                            |
| ---------------- | -------- | ------------------------------- |
| Content-Type     | 可选     | application/json; charset=utf-8 |

### 公共响应消息头

下表列出了所有声纹数据库接口的公共响应头域。HTTP 协议的标准响应头域不再这里列出。

| 消息头（Header） | 说明                                              |
| ---------------- | ------------------------------------------------- |
| Content-Type     | 只支持 JSON 格式，application/json; charset=utf-8 |

### 公共响应状态码

| 代码 | 含义         |
| ---- | ------------ |
| 200  | 请求成功     |
| 201  | 创建成功     |
| 400  | 请求参数错误 |
| 401  | 未授权       |
| 403  | 请求被拒绝   |
| 404  | 资源未找到   |
| 500  | 服务内部异常 |

### 响应数据格式

数据响应统一为如下格式。其中错误码取值来源公共内部状态码和各方法的专有内部状态码。

后续各接口只分别列出请求结果结构和专有内部状态码。

| 参数名 | 类型   | 说明           |
| ------ | ------ | -------------- |
| code   | number | 内部状态码     |
| msg    | string | 错误编码及描述 |
| data   | Object | 本次请求的结果 |

### 公共内部状态码

| 内部状态码 | 描述                                        |
| ---------- | ------------------------------------------- |
| 200          | 请求成功，结果保存在 data 字段中            |
| >200         | 错误编码，有错误发生，错误详情在 msg 字段中 |

## 1. HTTP API

> swagger链接： /api/data/v1/realtime/swagger

### 1.1 Micro服务相关

#### 1.1.1 查询服务健康状态

##### 接口功能

> 查看数据接收服务是否启动

##### URL

> /api/data/v1/realtime/health

##### 支持格式

> JSON

##### HTTP请求方式

> GET

##### 请求参数

> 无

##### 返回字段

> |返回字段|字段类型|说明                       |
> |:-----   |:------|:------------------------ |
> |health   |bool |返回结果状态。true：正常。    |

##### 接口示例

> 地址：/api/data/v1/realtime/health

``` javascript
{
    "health": true,
}
```

#### 1.1.2 查询服务信息

##### 接口功能

> 查询服务版本和配置信息

##### URL

> /api/data/v1/realtime/status

##### 支持格式

> JSON

##### HTTP请求方式

> GET

##### 请求参数

> 无

##### 返回字段

> |返回字段|字段类型|说明                       |
> |:-----   |:------|:------------------------ |
> |is_ok   |true |返回结果状态。true：正常。    |
> |basic   |json |基础配置信息    |
> |basic.zone     |string | 区域名                     |
> |basic.node     |string | 节点名                     |
> |basic.machine     |string | 服务器名称                     |
> |basic.service     |string | docker服务名称                     |
> |basic.instance     |string | docker实例ID号                     |
> |basic.app_name     |string | 服务名称                     |
> |basic.app_version     |string | 服务版本                     |
> |basic.app_root     |string | api接口地址                     |
> |basic.app_port     |string | 端口                     |
> |basic.is_dynamic_config     |bool | 是否使用动态配置                    |
> |basic.cpu_count     |int | 限制cpu数量                    |
> |basic.inswarm     |bool | 服务是否在集群中                     |
> |components     |json | 组件信息                     |
> |components.DataReceiverComponent     |json | 数据接收组件信息                     |
> |components.DataReceiverComponent.is_ok     |bool | 数据接收组件状态                     |
> |components.KVCache     |json | 集群KV同步组件信息                     |
> |components.KVCache.is_ok     |bool | 集群KV同步组件状态                     |
> |components.Kafka     |json | kafka组件信息                     |
> |components.Kafka.is_ok     |bool | kafka组件状态                     |
> |components.Logger     |json | 日志组件组件信息                     |
> |components.Logger.is_ok     |bool | 日志组件状态                     |
> |components.Taos     |json | 时序数据库组件信息                     |
> |components.Taos.is_ok     |bool | 时序数据库组件状态                     |
> |components.Trace     |json | 服务跟踪组件信息                     |
> |components.Trace.is_ok     |bool | 服务跟踪组件状态                     |

##### 接口示例

> 地址：/api/data/v1/realtime/status

``` javascript
{
  "is_ok": true,
  "basic": {
    "zone": "jnzone3",
    "node": "jnz3node1",
    "machine": "172-248",
    "service": "arc-consumer",
    "instance": "5ecc63e68547",
    "app_name": "data-receiver",
    "app_version": "v1.0.115",
    "api_root": "/api/data/v1/realtime",
    "api_port": 80,
    "is_dynamic_config": true,
    "cpu_count": -1,
    "inswarm": true
  },
  "components": {
    "DataReceiverComponent": {
      "is_ok": true
    },
    "KVCache": {
      "is_ok": true
    },
    "Kafka": {
      "is_ok": true
    },
    "Logger": {
      "is_ok": true
    },
    "Taos": {
      "is_ok": true
    },
    "Trace": {
      "is_ok": true
    }
  }
}
```

### 1.5 websocket接口

> 服务启动websocket服务，才能访问这个接口

#### 1.5.1 采集设备数据转发接口

##### 接口功能

> 建立长连接，接收音频数据，和声音强度，温度，震动波动图数据

##### URL

> /api/data/v1/realtime/ws/collector

##### 支持格式

> JSON

##### HTTP请求方式

> WEBSOCKET

##### 请求参数

- 连接成功发送数据

> |参数|必选|类型|说明|
> |:-----  |:-------|:-----|-----            |
> |collectorids   |true    |string|采集设备ID，多个ID用逗号分隔     |
> |interval   |true    |int|波动图时间间隔（毫秒）      |
> |temperature   |false    |bool|是否请求温度波动图数据      |
> |vibrate   |false    |bool|是否请求震动原始数据      |
> |vibrate_vatd   |false    |bool|是否请求震动波动图数据      |
> |audio   |false    |bool|是否请求音频原始数据      |
> |audio_vatd   |false    |bool|是否请求音频强度波动图数据      |
> |audio_spark |false |bool|是否请求音频电火花波动图数据 |
> |fill   |false    |bool|是否填充数据      |

##### 返回字段

- 音频原始数据为二进制数据
- 振动原始数据为二进制数据
- 波动图数据

> |参数|类型|说明|
> |:-----  |:-----|-----            |
> |temperature   |json|温度波动图数据     |
> |temperature.collectorid   |string|采集设备ID     |
> |temperature.data   |array|温度数据     |
> |temperature.data[0].time     |int | 时间点(毫秒)|
> |temperature.data[0].t     |float | 温度值|
> |vibrate   |json|震动波动图数据     |
> |vibrate.collectorid   |string|采集设备ID     |
> |vibrate.samplerate   |int|震动采样率     |
> |vibrate.range   |int|震动量程范围     |
> |vibrate.resolution   |int|震动分辨率     |
> |vibrate.offset_x   |int|x轴偏移     |
> |vibrate.offset_y   |int| y轴偏移     |
> |vibrate.offset_z   |int|z轴偏移      |
> |vibrate.data   |array|震动数据     |
> |vibrate.data[0].time     |int | 时间点(毫秒)|
> |vibrate_engine   |json|振动过算法引擎波动图数据     |
> |vibrate_engine.collectorid   |string|采集设备ID     |
> |vibrate_engine.data   |array|振动过算法引擎数据     |
> |vibrate_engine.data[0].time     |int | X轴时间点(毫秒)|
> |vibrate_engine.data[0].acc     |float | X轴加速度 |
> |vibrate_engine.data[0].rms     |float | X轴均方根（强度）|
> |vibrate_engine.data[0].varance     |float | X轴方差|
> |vibrate_engine.data[0].std     |float | X轴标准差|
> |vibrate_engine.data[0].skew     |float | X轴偏度|
> |vibrate_engine.data[0].raw     |float | X轴原始采样值|
> |vibrate_engine.data[0].ppv     |float | X轴峰峰值|
> |vibrate_engine.data[0].peak     |float | X轴峰值|
> |vibrate_engine.data[0].minv     |float | X轴最小值|
> |vibrate_engine.data[0].maxv     |float | X轴最大值|
> |vibrate_engine.data[0].mean     |float | X轴均值|
> |vibrate_engine.data[0].kurt     |float | X轴峭度|
> |vibrate_engine.data[0].ind_waveform     |float | X轴波形指标|
> |vibrate_engine.data[0].ind_peak     |float | X轴峰值指标|
> |vibrate_engine.data[0].ind_margin     |float | X轴裕度指标|
> |vibrate_engine.data[0].ind_kurt     |float | X轴峭度指标|
> |vibrate_engine.data[0].ind_impulse     |float | X轴脉冲指标|
> |vibrate_engine.data[1].time     |int | Y轴时间点(毫秒)|
> |vibrate_engine.data[1].acc     |float | Y轴加速度 |
> |vibrate_engine.data[1].rms     |float | Y轴均方根（强度）|
> |vibrate_engine.data[1].varance     |float | Y轴方差|
> |vibrate_engine.data[1].std     |float | Y轴标准差|
> |vibrate_engine.data[1].skew     |float | Y轴偏度|
> |vibrate_engine.data[1].raw     |float | Y轴原始采样值|
> |vibrate_engine.data[1].ppv     |float | Y轴峰峰值|
> |vibrate_engine.data[1].peak     |float | Y轴峰值|
> |vibrate_engine.data[1].minv     |float | Y轴最小值|
> |vibrate_engine.data[1].maxv     |float | Y轴最大值|
> |vibrate_engine.data[1].mean     |float | Y轴均值|
> |vibrate_engine.data[1].kurt     |float | Y轴峭度|
> |vibrate_engine.data[1].ind_waveform     |float | Y轴波形指标|
> |vibrate_engine.data[1].ind_peak     |float | Y轴峰值指标|
> |vibrate_engine.data[1].ind_margin     |float | Y轴裕度指标|
> |vibrate_engine.data[1].ind_kurt     |float | Y轴峭度指标|
> |vibrate_engine.data[1].ind_impulse     |float | Y轴脉冲指标|
> |vibrate_engine.data[2].time     |int | Z轴时间点(毫秒)|
> |vibrate_engine.data[2].acc     |float | Z轴加速度 |
> |vibrate_engine.data[2].rms     |float | Z轴均方根（强度）|
> |vibrate_engine.data[2].varance     |float | Z轴方差|
> |vibrate_engine.data[2].std     |float | Z轴标准差|
> |vibrate_engine.data[2].skew     |float | Z轴偏度|
> |vibrate_engine.data[2].raw     |float | Z轴原始采样值|
> |vibrate_engine.data[2].ppv     |float | Z轴峰峰值|
> |vibrate_engine.data[2].peak     |float | Z轴峰值|
> |vibrate_engine.data[2].minv     |float | Z轴最小值|
> |vibrate_engine.data[2].maxv     |float | Z轴最大值|
> |vibrate_engine.data[2].mean     |float | Z轴均值|
> |vibrate_engine.data[2].kurt     |float | Z轴峭度|
> |vibrate_engine.data[2].ind_waveform     |float | Z轴波形指标|
> |vibrate_engine.data[2].ind_peak     |float | Z轴峰值指标|
> |vibrate_engine.data[2].ind_margin     |float | Z轴裕度指标|
> |vibrate_engine.data[2].ind_kurt     |float | Z轴峭度指标|
> |vibrate_engine.data[2].ind_impulse     |float | Z轴脉冲指标|
> |audio_engine   |json|音频过算法引擎波动图数据     |
> |audio_engine.collectorid   |string|采集设备ID     |
> |audio_engine.data   |array|音频过算法引擎数据     |
> |audio_engine.data[0].time     |int | 时间点(毫秒)|
> |audio_engine.data[0].db     |float | 强度值分贝 |
> |audio_engine.data[0].rms     |float | 均方根（强度）|
> |audio_engine.data[0].varance     |float | 方差|
> |audio_engine.data[0].std     |float | 标准差|
> |audio_engine.data[0].skew     |float | 偏度|
> |audio_engine.data[0].raw     |float | 原始采样值|
> |audio_engine.data[0].ppv     |float | 峰峰值|
> |audio_engine.data[0].peak     |float | 峰值|
> |audio_engine.data[0].minv     |float | 最小值|
> |audio_engine.data[0].maxv     |float | 最大值|
> |audio_engine.data[0].mean     |float | 均值|
> |audio_engine.data[0].kurt     |float | 峭度|
> |audio_engine.data[0].ind_waveform     |float | 波形指标|
> |audio_engine.data[0].ind_peak     |float | 峰值指标|
> |audio_engine.data[0].ind_margin     |float | 裕度指标|
> |audio_engine.data[0].ind_kurt     |float | 峭度指标|
> |audio_engine.data[0].ind_impulse     |float | 脉冲指标|
> |audio_engine.data[0].spark |float | 音频电火花数值 |
> |audio   |json|音频数据信息     |
> |audio.collectorid   |string|采集设备ID     |
> |audio.samplerate   |int|采样率      |
> |audio.bits   |int|位深度      |
> |audio.channel   |int|通道数      |
> |audio.data   |array|音频二进制数据，这个数据单独发送，不与json数据在一起      |
> |audio.data[0].time     |int | 时间点(毫秒)|

##### 接口示例

> 地址：/api/data/v1/realtime/ws/collector

 - 客户端发送

``` javascript
{
    "collectorids": "A00000000000",
    "interval": 100,
    "temperature": true,
    "vibrate": true,
    "audio_vatd": true,
    "audio_spark": true,
    "audio": true,
    "fill": true, // 采集设备在interval时间内没有数据，就会发送默认填充数据
}
```

 - 服务端返回音频数据(采集设备发送数据)

``` javascript
{
    "audio": {
        "collectorid": "A00000000000",
        "samplerate": 32000,
        "bits": 16,
        "channel": 1
        "data" [{
            "time": 1615178437100,
        }]
    }
}
```

 - 服务端返回音频数据(采集设备发送数据, 每个数据包都发送)

``` javascript
// 二进制pcm数据
```

 - 服务端返回振动数据(采集设备发送数据)

``` javascript
{
    "vibrate": {
        "collectorid": "A00000000000",
        "samplerate":3200,
        "range": 16,
        "resolution": 13,
        "offset_x": 0,
        "offset_y": 0,
        "offset_z": 0,
        "data" [{
            "time": 1615178437100,
        }]
    }
}
```

 - 服务端返回振动数据(采集设备发送数据, 每个数据包都发送)

``` javascript
// 二进制pcm数据
```

 - 服务端返回(采集设备发送数据, 每隔interval时间)

``` javascript
{
    "temperature": {
        "collectorid": "A00000000000",
        "data" [{
            "time": 1615178437100,
            "t": 32.828
        }]
    },
    "vibrate_engine": {
        "collectorid": "A00000000000",
        "data" [{
			"time": 1615178437100,  	// - 数据保存时间(毫秒时间戳)
			"ind_impulse": 4.6229773,  	// - 脉冲指标
			"ind_kurt": 3.4695735,		// - 峭度指标
			"ind_margin": 7.070554,		// - 裕度指标
			"ind_peak": 3.01643, 		// - 峰值指标
			"ind_waveform": 1.532599,	// - 波形指标
			"kurt": 0.74414146,			// - 峭度
			"maxv": 2.0527596,			// - 最大值
			"mean": -0.05275969,		// - 均值
			"minv": -1.9472404,			// - 最小值
			"peak": 2.0527596,			// - 峰值
			"ppv": 4,					// - 峰峰值
			"raw": 1,					// - 原始采样值
			"rms": 36.8288346877526,	// - 均方根（强度）
			"skew": -0.020787,			// - 偏度
			"std": 0.67847794,			// - 标准差
			"varance": 0.4603323,		// - 方差
			"acc": 1.8288346877526,	    // - 能量
        },{
			"time": 1615178437100,  	// - 数据保存时间(毫秒时间戳)
			"ind_impulse": 4.6229773,  	// - 脉冲指标
			"ind_kurt": 3.4695735,		// - 峭度指标
			"ind_margin": 7.070554,		// - 裕度指标
			"ind_peak": 3.01643, 		// - 峰值指标
			"ind_waveform": 1.532599,	// - 波形指标
			"kurt": 0.74414146,			// - 峭度
			"maxv": 2.0527596,			// - 最大值
			"mean": -0.05275969,		// - 均值
			"minv": -1.9472404,			// - 最小值
			"peak": 2.0527596,			// - 峰值
			"ppv": 4,					// - 峰峰值
			"raw": 1,					// - 原始采样值
			"rms": 36.8288346877526,	// - 均方根（强度）
			"skew": -0.020787,			// - 偏度
			"std": 0.67847794,			// - 标准差
			"varance": 0.4603323,		// - 方差
			"acc": 1.8288346877526,	    // - 能量
        },{
			"time": 1615178437100,  	// - 数据保存时间(毫秒时间戳)
			"ind_impulse": 4.6229773,  	// - 脉冲指标
			"ind_kurt": 3.4695735,		// - 峭度指标
			"ind_margin": 7.070554,		// - 裕度指标
			"ind_peak": 3.01643, 		// - 峰值指标
			"ind_waveform": 1.532599,	// - 波形指标
			"kurt": 0.74414146,			// - 峭度
			"maxv": 2.0527596,			// - 最大值
			"mean": -0.05275969,		// - 均值
			"minv": -1.9472404,			// - 最小值
			"peak": 2.0527596,			// - 峰值
			"ppv": 4,					// - 峰峰值
			"raw": 1,					// - 原始采样值
			"rms": 36.8288346877526,	// - 均方根（强度）
			"skew": -0.020787,			// - 偏度
			"std": 0.67847794,			// - 标准差
			"varance": 0.4603323,		// - 方差
			"acc": 1.8288346877526,	    // - 能量
        }]
    }
    "audio_engine": {
        "collectorid": "A00000000000",
        "data": [{
			"time": 1615178437100,  	// - 数据保存时间(毫秒时间戳)
			"ind_impulse": 4.6229773,  	// - 脉冲指标
			"ind_kurt": 3.4695735,		// - 峭度指标
			"ind_margin": 7.070554,		// - 裕度指标
			"ind_peak": 3.01643, 		// - 峰值指标
			"ind_waveform": 1.532599,	// - 波形指标
			"kurt": 0.74414146,			// - 峭度
			"maxv": 2.0527596,			// - 最大值
			"mean": -0.05275969,		// - 均值
			"minv": -1.9472404,			// - 最小值
			"peak": 2.0527596,			// - 峰值
			"ppv": 4,					// - 峰峰值
			"raw": 1,					// - 原始采样值
			"rms": 36.8288346877526,	// - 均方根（强度）
			"skew": -0.020787,			// - 偏度
			"std": 0.67847794,			// - 标准差
			"varance": 0.4603323,		// - 方差
			"spark": 1.8288346877526,	// - 电火花事件识别值
            "db": 1.8288346877526	    // - 强度
        }]
    }
}
```
