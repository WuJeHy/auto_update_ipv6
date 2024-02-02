# alidns ipv6 动态更新 工具


因为有某些机器需要定时更新ipv6 地址 

故 写个程序定时更新本机的公网ip到指定的域名下

基于 ip 命令行工具解析地址

使用语言 ： golang 

update 20231122: 
    因为业务升级搞了个服务器,使用了vm系统,为了管理实例化后的机器对应的ipv6 , 增加了对esxi 的支持
    可以将指定实例的 地址刷新到 阿里云. 

update 20240202:
    因为一些原因 旧版本发现了一些bug 以及一些 管理麻烦的问题 
    于是整了个交互页面, 本来想 做 tui 来管理, 然后太麻烦了, 直接用命令吧 , 简单粗暴
    

# 编译 
 

ubuntu 22.10 
```shell
make build_ipv6_mgr_ubuntu
```

 

## 高级(旧)

核心逻思路

因为通过某写域名查询当前ip 不准确 所以使用 ip 命令解析 

同时ip 命令有有效期等信息 可以更好的选择需要的ip 

核心代码在  [GetTargetIPv6Info](tools/ipv6.go ) 方法

## 设计思路
    首先有一个管理器 进行 配置管理 
    然后有个定时器 定期探测 
    然后 有个 rpc 进行外部管理
    因为 不想做鉴权 , 所以直接使用  unix socket 默认同级目录访问
    然后提供 一个 参数解析 实现命令简单交互 


### 配置：

新版本有两个配置 
一个系统配置 
[config.yaml](config.yaml)

详细配置文件: 
[config.go](manager%2Fconfig.go)

的 ConfigFile 结构

```yaml
#日志配置
#默认  logs 目录  info等级
logger:
  logger_path: logs
  logger_level: info

aliyun:
  endpoint: "alidns.xxxxxx.aliyuncs.com"
  access_key_id: "-" # 阿里云 key id
  access_key_secret: "-" # 阿里云 key id
app:

  update_wait_time: 120
  net_name: "wlp5s0" # ip addr 需要查询ipv6 的接口
# 设置 记录储存的 配置文件 支持 yaml json
# 空则默认 record.yaml 
records_file: "my_record.yaml"

# 设置 unix sock 路径
# 空则 默认 当前路径 pio_grpc_ddns.sock
unix_sock: "pio_grpc_ddns.sock"
# esxi 相关配置
esxi:
  url: "https://192.168.1.2" # esxi 的 后台地址
  username: "root"
  password: "1234678"
  insecure: true
```



### 运行

注意 : 简单粗暴 执行操作默认 会后台启动 实例 

提示: 配置文件不会自动刷新 ,需要 `-store` 参数完成 , 并且 `-store` 可以并行处理

修改完需要保存的时候一定要 `-store`
修改完需要保存的时候一定要 `-store`
修改完需要保存的时候一定要 `-store`

参数信息 默认读取但前路径下的 config.yaml 的配置信息

日志默认按小时 保存到 logs 下

```shell

./mgr_2024020219-amd64_ubuntu -h           
Usage of ./mgr_2024020219-amd64_ubuntu:
  -add
        添加一个记录
  -bg
        后台运行
  -c string
        配置文件 (default "./config.yaml")
  -del
        删除一个记录
  -logger_level
        更新日志等级
  -reload
        重载记录配置
  -show
        查看记录
  -stdout string
        stdout 重定向 (default "mgr.out")
  -store
        保存记录配置
  -ui
        启动ui
  -unix string
        unix sock 文件 路径
  -update
        更新记录

```

```shell
./mgr_2024020219-amd64_ubuntu -show        
2024/02/02 19:30:05 新任务ID: 3342645
2024/02/02 19:30:06 链接成功
序号:             RecordID          RR      Type               WatchType        VMName
-----------------------------------------------------------------------------------------------
0000:    ----------------          xx      AAAA          WatchTypeLocal
0001:    ----------------         xxx      AAAA          WatchTypeLocal
0002:    ----------------         xxx      AAAA          WatchTypeLocal
-----------------------------------------------------------------------------------------------
记录总数: 3

```


其他, 略



