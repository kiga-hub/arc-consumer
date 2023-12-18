# 单元测试 


# 版本信息记录

| 日期     | 版本 | 说明 | 作者 |
| -------- | ---- | ---- | ---- |
| 20220322 | 1    |      | 余琦     |


## 目的

> 测试各个模块的功能是否符合需求


## 术语和定义


## 测试整体设计

> 采用Go Convey + Go Monkey 的单元测试工具


## 测试方法

> 命令行执行：go test -v 或 go test -cover -v [指定测试文件]

#测试模块

## 1. pkg/cache

### 1.1 测试功能
> cache：audio/temperature/vibrate 根据id存储数据，获取数据

### 1.2 流程图
```mermaid
    graph TB
    Begin(开始) --> Set[设置id及默认配置]
    Set --> New[建立连接对象并开启服务] --> SetTime[设置存储数据设置时间]
    SetTime --> Saudio[audio] --> AInput[AInput] --> Input[Input]
    SetTime --> Stemperature[temperature] --> TInput[TInput] --> Input[Input]
    SetTime --> Svibrate[vibrate] --> VInput[VInput] --> Input[Input]
    Input --> |存储| Internal[Internal]
   	Input --> Return[return存储值] --> Sif[判断是否存储成功]
    Set --> Offset[设置偏移时间及时间段] -->Get[根据id取值]
    Get --> Gaudio[audio] --> ASearch[ASearch] --> Search[Search]
    Get --> Gtemperature[temperature] --> TSearch[TSearch] --> Search[Search]
    Get --> Gvibrate[vibrate] --> VSearch[VSearch] --> Search[Search]
    Search --> |获取| Internal[Internal]
    Search --> Greture[reture存储值] --> encode[解码/转换类型] --> Gif[判断缓存值与获取值是否一致] --> Close[G结束]
```
![](./cache.png)

### 1.3 测试需求及结果
| 序号 | 需求 | 输入 | 输出 | 是否满足需求 | 覆盖率 | 说明 |
| --- | ----- | ---- | ---- | ---- | ---- | ---- |
| 1 | 执行audio | []byte{1, 2, 3} | []byte{1, 2, 3} | true | 51.8% | 存储byte类型取出还是byte类型 |
| 2 | 执行temperature | 25 | 25 | true | 51.8% | 存储int16类型，返回float32类型，需要转换类型 |
| 2 | 执行vibrate | {0 0 0} | {0 0 0} | true | 51.8% | 存储结构体类型，获取了处理后进行比较 |


## 2. pkg/proxy

### 2.1 测试功能
> proxy：tcp/udp发送数据是否成功

### 2.2 流程图
```mermaid
    graph TB
    Begin(开始) --> Set[设置默认配置]
    Set --> New[New创建连接对象]
    Set--> Routine[开启tcp/udp服务协程]
    Routine --> Listener[监听端口]
    New --> Start[Start开启连接服务] --> Sleep[暂停2s触发执行任务] --> Write[构造数据并Write发送]
    Write -->|发送数据| Listener
    Listener -->|返回数据长度| Write
    Write --> So[发送数据长度与返回是否一致判断是否发送成功]
   	So --> send[退出程序并结束服务协程] --> close[结束]
   	So --> |结束goroutine| Listener
```
![](./proxy.png)

### 2.3 测试需求及结果
| 序号 | 需求 | 输入 | 输出 | 是否满足需求 | 覆盖率 | 说明 |
| --- | ----- | ---- | ---- | ---- | ---- | ---- |
| 1 | 执行tcp New Start Write无错误 write成功返回输入参数长度 | 123456 | 6 | true | 90.0% |  |
| 2 | 执行udp New Start Write无错误 write成功返回输入参数长度 | 654321 | 6 | true | 90.0% |  |


