# alidns ipv6 动态更新 工具


因为有某些机器需要定时更新ipv6 地址 

故 写个程序定时更新本机的公网ip到指定的域名下

基于 ip 命令行工具解析地址

使用语言 ： golang 

update 20231122: 
    因为业务升级搞了个服务器,使用了vm系统,为了管理实例化后的机器对应的ipv6 , 增加了对esxi 的支持
    可以将指定实例的 地址刷新到 阿里云. 

# 编译 


linux 静态编译

```shell
make build_ipv6_app

```

ubuntu 22.10 
```shell
make build_ipv6_app_ubuntu
```


以上编译区别是ldd 是否有外部依赖

```text

(base) ➜  auto_update_ipv6 git:(master) ✗ ldd build/ipv6_ddns_2023070618-amd64_*
build/ipv6_ddns_2023070618-amd64_linux:
        not a dynamic executable
build/ipv6_ddns_2023070618-amd64_ubuntu:
        linux-vdso.so.1 (0x00007ffdaa7dd000)
        libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007fb0e320e000)
        libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007fb0e2e00000)
        /lib64/ld-linux-x86-64.so.2 (0x00007fb0e323a000)
```


## 高级

核心逻思路

因为通过某写域名查询当前ip 不准确 所以使用 ip 命令解析 

同时ip 命令有有效期等信息 可以更好的选择需要的ip 

核心代码在  [GetTargetIPv6Info](tools/ipv6.go ) 方法



### 配置：

```yaml
ddns:
  endpoint: "alidns.cn-shenzhen.aliyuncs.com"
  access_key_id: "-" # 阿里云 key id
  access_key_secret: "-" # 阿里云 key id
  net_name: "wlp5s0" # 网卡位置
  update_wait_time: 120 # 轮讯检测时间
  retry_time: 300 # 重试时间
  esxi:
    url: "https://192.168.1.2" # esxi 的 后台地址
    username: "root"
    password: "1234678"
    insecure: true
  records: # 需要更新的目标
    -
      RR: "test"  # 二级域名
      Type: "AAAA" # 记录值 ipv6 是AAAA
      RecordId: "---" # 记录id

    -
      RR: "test2"
      Type: "AAAA"
      RecordId: "---"
      WatchType: "local" # 标记为 来源于本机的ipv6 , 不填默认是本机
    - RR: "test3"
      Type: "AAAA"
      RecordId: "----"
      WatchType: "esxi" # 标记来源于 esxi 的 ipv6
      VMName: "vm_name" # 对应 esxi 的 实例名
```



### 运行

如下直接后台运行

```shell
./ipv6_ddns_2023112218-amd64_linux -bg 
```

参数信息 默认读取但前路径下的 config.yaml 的配置信息

```
./build/ipv6_ddns_2023070618-amd64_linux -h
Usage of ./build/ipv6_ddns_2023070618-amd64_linux:
  -bg
        后台运行
  -c string
        配置文件 (default "./config.yaml")
  -stdout string
        stdout 重定向 (default "auto_dns.out")

```

日志默认按小时 保存到 logs 下



