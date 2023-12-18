# 项目文档

- 后端：用`echo`快速搭建基础restful风格API。
- 网络服务器框架：使用`gnet`网络框架实现数据接收。
- 数据库：采用`TDEngine`(2.2.2.0)版本，使用`taosSql`实现对数据库的基本操作。
- API文档：使用`Swagger`构建自动化文档。
- 配置文件：使用`viper`解析配置文件。
- CLI: 使用`cobra`实现命令行参数。
- grpc: 转发数据

## 1. 基本介绍

### 1.1 项目介绍

> arc-consumer是一个基于gnet框架开发的后台数据接收服务，集成数据缓存(cache)，数据存储(file、taos)，数据转发(kafka、grpc、tcp)，序列化波动图(websocket)，模拟采集设备等功能。

## 2. 安装说明

- golang版本 >= v1.20
- IDE推荐：Goland