package main

import (
	"flag"
	_ "net/http/pprof"
	"quick/initialize"
	"quick/router"
)

func main() {
	conf := *flag.String("f", "./config/config.toml", "config file")
	flag.Parse()
	initialize.Initialize(conf)
	router.Run()
}
