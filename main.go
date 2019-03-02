package main

/*
   1、需要添加热加载功能
   2、需要添加读取配置文件的功能，配置文件中需要添加日志的路径，还有etcd的地址
   3、要往etcd中添加服务的地址，路由等信息
    获取情况如下：
		map["路由"]服务的地址ip的Slice？
*/

import (
	"quick/router"
	_ "net/http/pprof"
	"quick/initialize"
	"flag"
)


func main(){
	conf := *flag.String("f", "./config/config.toml", "config file")
	flag.Parse()
	initialize.Initialize(conf)
	router.Run()
}
