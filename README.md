# README

<img src='https://github.com/EpochCloud/quick/blob/master/png/logo.png' div align=center; />

## 简介

quick是一款专为微服务架构定制的高性能网关


##  特点
- 灰度发布
- 负载均衡
- 智能路由
- 熔断
- 轻量级
- 轻松任何基础架构，可以部署在云或者云内部部署，也可直接在物理机运行
- 跨语言，所有语言都可以使用
- 可作为分布式网关、微服务网关
  ...


## 首先运行confcenter
1. 首先注册quick的配置信息

  - /quick_operation  ：注册quick的配置

  - /gateway_configuration ：需要注册的服务


2. go get github.com/EpochCloud/quick
    这里go get的时候可能会报如下信息：

  ```shell
package ConfCenter/initialization: unrecognized import path "ConfCenter/initialization" (import path does not begin with hostname)
package ConfCenter/router: unrecognized import path "ConfCenter/router" (import path does not begin with hostname)
  ```

​     ***这是我没适配github路径的问题，暂时不用管***



3. 在quick项目的config文件夹中，有一个config.toml实例配置文件
   只需要把前面的ip和port更改为confcenter的ip和port就可以
   <b>比如：</b>

   - clone代码下来的toml文件内容为

   ```toml
   [confCenter]
       Addr = "http://127.0.0.1:8081/quick_configuration"
       SrvAddr = "http://127.0.0.1:8081/quick_operation"
   ```

   - 如果confcenter在192.168.51.11机器上面，那么只需要修改成如下即可

   ```toml
   [confCenter]
      Addr = "http://192.168.51.11:8081/quick_configuration"
      SrvAddr = "http://192.168.51.11:8081/quick_operation"
   ```


> **注意**：这里的toml有效只有一次，也就是刚启动的时候有效，启动之后这个配置文件就没什么用了，所以不用担心，以后不会修改这个配置文件
> 即使confcenter从192.168.51.11:8081迁移到192.168.51.12:8082 这也是没有问题的，quick会依旧保持高性能的运行状态，没有任何影响


## 运行quick

 在此步骤之前运行了confcenter并且注册了信息，方可进行下面步骤
```shell
cd $GOPATH
go get github.com/EpochCloud/quick
mv ./src/github.com/EpochCloud/quick ./src/
go build quick
mv quick src/quick/
```

win 环境
```shell
quick.exe
```

linux环境
```shell
./quick
```

## 示例

注意由于本人能力有限，没有给出前端代码，所以下面一切测试均在postman中测试

  ***目前正在积极开发confcenter管理平台ConfCenter_web_Admin，有些功能已经可用，具体链接如下 https://github.com/EpochCloud/ConfCenter_web_Admin   具体按照相应Readme进行安装，可以使用此平台来替代postman***

1、运行confcenter并且注册quick的配置
```go
/gateway_configuration  //注册quick服务  post
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
```

2、运行quick

3、注册服务到网关

```go
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
```

4、编译安装
```shell
go github.com/EpochCloud/gin_example_for_quick
go install github.com/EpochCloud/gin_example_for_quick
cd bin
```

win环境
```powershell
gin_example_for_quick.exe
```

linux\mac
```shell
./gin_example_for_quick
```

5、然后直接访问quick的外网
http://127.0.0.1:8090/  get请求
返回值为 hello word


## 灰度发布


在此之前一定已经完成了示例部分，现在已经对quick比较熟悉了，下面来尝试一下quick的灰度发布功能
要灰度发布一个新的路由

示例
1、打开postman
```go
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
```
2、访问quick的domain + / 便可以访问到这个要灰度的服务


## 紧急修复、下线


如果在线上遇到了某个路由或者某个服务有bug，要下线，那么这种情况是可以轻松实现的

示例
1、打开postman
```go
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
```
2、访问quick的domain + / 便访问不到到这个下线的服务


## 注意

1. 使用灰度和紧急修复下线都是秒级进行的，不会存在卡顿，所以不用担心使用这些功能会出现网关卡顿等对业务的影响
2. 使用紧急下线功能是对已有的链接还是正常处理，只是不再有新的请求到下线的服务，所以不用担心下线之后以前的链接没有处理完毕就中断导致的问题


## 启动quick可能遇到的问题

```shell
loglevel is not null
```
这样需要你继续回到上面的**示例1**，把下面的logpath换成自己的
```go
/gateway_configuration  //post
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
```

