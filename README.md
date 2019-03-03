# README

## 简介

```go
quick是一款专为微服务架构定制的高性能网关
特点：
灰度发布
负载均衡
智能路由
熔断
轻量级
轻松任何基础架构，可以部署在云或者云内部部署，也可直接在物理机运行
跨语言，所有语言都可以使用
可作为分布式网关、微服务网关
...
```

## 首先运行confcenter

```go

1、首先注册quick的配置信息
/quick_operation  ：注册quick的配置
/gateway_configuration ：需要注册的服务
2、在quick的config的config.toml文件中注册
只需要把前端的ip和port更改为confcenter的ip和port就可以
比如：
   clone代码下来的toml文件内容为
   [confCenter]
    Addr = "http://127.0.0.1:8081/quick_configuration"
    SrvAddr = "http://127.0.0.1:8081/quick_operation"
如果confcenter在192.168.51.11机器上面，那么只需要修改成如下即可
 [confCenter]
    Addr = "http://192.168.51.11:8081/quick_configuration"
    SrvAddr = "http://192.168.51.11:8081/quick_operation"

注意：这里的toml有效只有一次，也就是刚启动的时候有效，启动之后这个配置文件就没什么用了，所以不用担心，以后不会修改这个配置文件
即使confcenter从192.168.51.11:8081迁移到192.168.51.12:8082 这也是没有问题的，quick会依旧保持高性能的运行状态，没有任何影响

```



## 运行quick

```go
 在此步骤之前运行了confcenter并且注册了信息，方可进行下面步骤
  cd $GOPATH
  cd src/github.com/EpochCloud/quick
  go install 
  cd $GOPATH
  win 环境
  quick.exe -f ./src/quick/config/config.toml
  linux环境
  ./quick -f ./src/quick/config/config.toml
```



## 示例

```go
注意由于本人能力有限，没有给出前端代码，所以下面一切测试均在postman中测试

1、运行confcenter并且注册quick的配置
    /gateway_configuration  //注册quick服务
    {
        "ip":"127.0.0.1",     //quick的外网地址
        "port":"8090",        //quick的外网端口
        "timeout":15,         //quick的平滑重启超时时间
        "loglevel":"debug",   //quick的日志级别
        "logpath":"D:/project/src/quick/logcatlog",  //quick打印日志的路径
        "bufpool":0,          //需要的缓存池数量，默认0
        "intranetip":"127.0.0.1",  //内网的ip
        "intranetport":"6060",     //内网的端口    
        "managerroute":"/manager",  //quick内网的配置路由
        "serviceroute":"/service"   //quick的内网服务路由
    }
    
2、运行quick
3、注册服务到网关
	/quick_operation  //注册接入服务   
    {
        "route": "/",    //服务的路由
        "service": {
            "serviceaddr": ["127.0.0.1:6067"],  //服务的地址
            "registertime": "2019.2.22",   //注册时间
            "altreason": "patch test",     //注册原因
            "balance":"random"             //选择这个路由的负载均衡策略
        },
        "servicename": "liantiao"          //注册的服务名字
    }
4、git clone gin_example_for_quick,并且启动
5、这样直接访问quick的外网
http://127.0.0.1:8090/  get请求
返回值为 hello word
```

## 灰度发布

```go
在此之前一定已经完成了示例部分，现在已经对quick比较熟悉了，下面来尝试一下quick的灰度发布功能
要灰度发布一个新的路由

示例
1、打开postman
domain  ：confcenter的地址
method  ：POST
route   ：/quick_operation    
    {
        "route": "/",    //选择要灰度的路由
        "service": {
            "serviceaddr": ["127.0.0.1:6067"],  //灰度的地址
            "registertime": "2019.2.22",   //灰度时间
            "altreason": "patch test",     //灰度原因
            "balance":"random"             //灰度这个路由的负载均衡策略
        },
        "servicename": "liantiao"          //灰度服务名字
    }
2、访问quick的domain + / 便可以访问到这个要灰度的服务
```

## 紧急修复、下线

```go
如果在线上遇到了某个路由或者某个服务有bug，要下线，那么这种情况是可以轻松实现的

示例
1、打开postman
domain  ：confcenter的地址
method  ：delete
route   ：/quick_operation    
    {
        "route": "/",    //选择要下线的路由
        "service": {
            "serviceaddr": ["127.0.0.1:6067"],  //下线的地址
            "registertime": "2019.2.22",   //下线原因
            "balance":"random"             //下线这个路由的负载均衡策略
        },
        "servicename": "liantiao"          //下线服务名字
    }
2、访问quick的domain + / 便访问不到到这个下线的服务
```

## 注意

```go
使用灰度和紧急修复下线都是秒级进行的，不会存在卡顿，所以不用担心使用这些功能会出现网关卡顿等对业务的影响

```

