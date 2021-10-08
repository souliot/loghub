# loghub 日志采集及系统监控服务

主要实现步态系统日志集中采集及操作系统监控

> 日志存储 clickhouse
> 系统监控暴露 http 接口，数据格式采用 prometheus

## 框架

> 主要包括 config | models | srv | utils

### 配置 config

> 服务配置初始化  
> 先加载默认配置，再读取配置文件，合并配置。

### 核心模块 models

logcollect | metrics | server | ws

> server：基于 gin，实现 http server  
> logcollect：日志采集功能，input(日志文件)，output（clickhouse）  
> metrics：系统监控信息  
> ws：暴露 websocket 接口，可以通过 websocket 查看服务实时日志。

### 服务接口 srv

> 实现 servicelib 接口，用于服务注册初始化配置中心配置等工作。  
> srv/version.go 实现版本升级功能

### 工具类 utils

livego 用到的基础工具类包
