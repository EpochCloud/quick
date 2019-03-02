package router

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"quick/config"
	"quick/log"
	"sync"
	"syscall"
	"time"
)

func Run() {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go run(wg)
	go Intranet(wg)
	wg.Wait()
}

func run(wg *sync.WaitGroup) {
	domain := fmt.Sprintf("%s:%s", config.M.OldGateway.Ip, config.M.OldGateway.Port)
	defer func() {
		if err := recover(); err != nil {
			log.Error("%v goroutine err:%v", domain, err)
		}
	}()
	server := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
		Addr:         domain,
		Handler:      &config.S{},
	}
	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	go shutdown(exit, wg, server, domain)
	config.Log.Debug("[%s] the http server run %v", time.Now(), domain)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("[%v] the http %v run err %v", time.Now(), domain, err)
	}
}

func Intranet(wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("localhost:6666 goroutine err", err)
		}
	}()
	inDomain := fmt.Sprintf("%s:%s", config.M.OldGateway.IntranetIp, config.M.OldGateway.IntranetPort)
	log.Debug("config.M.OldGateway intraneIp is :%v", inDomain)
	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
		Addr:         inDomain,
		Handler:      &config.Manager{},
	}
	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	go shutdown(exit, wg, srv, inDomain)
	log.Debug("the Intranet http server run localhost:6060")
	config.Log.Debug("[%s] the Intranet http server run localhost:6060", time.Now())
	srv.ListenAndServe()
}

func shutdown(exit chan os.Signal, wg *sync.WaitGroup, srv *http.Server, domain string) {
	<-exit
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.M.OldGateway.TimeOut)*time.Second)
	defer func() {
		cancel()
		close(config.ServiceChan)
		close(config.ManagerChan)
		wg.Done()
	}()
	log.Warn(" gracefully shutdown the http server %s", domain)
	err := srv.Shutdown(ctx)
	if err != nil {
		config.Log.Error("http server %v shutdown err:%v", domain, err)
		return
	}
}
