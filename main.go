package main

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
