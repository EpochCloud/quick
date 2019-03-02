package initialize

import (
	"bytes"
	"encoding/json"
	"github.com/BurntSushi/toml"
	"github.com/DeanThompson/syncmap"
	"io/ioutil"
	"net/http"
	"quick/balance"
	s "quick/basic/app"
	"quick/config"
	"quick/engine"
	"quick/log"
	"sync"
	"time"
	"errors"
)

type Initialization struct {
	bufPool, client *sync.Pool
	conf            conf
}

type conf struct {
	ConfCenter confCenter
}

type confCenter struct {
	Addr    string
	SrvAddr string
}

func init() {
	config.M = config.NewManagers()
	config.Srv = config.NewS()
	config.ManagerChan = make(chan string, 10)
	config.ServiceChan = make(chan string, 10)
	config.DeleteChan = make(chan string, 10)
	config.InsertChan = make(chan string, 10)
}

func Initialize(conf string) {
	newInitialization := NewInitialization().initConfig(conf).do().logInitialize().bufPoolBasic().serverClient().pullServer().Reload()
	config.Buf = newInitialization.bufPool
	config.Client = newInitialization.client
}

func NewInitialization() *Initialization {
	return &Initialization{}
}

func (initialization *Initialization) logInitialize() *Initialization {
	if config.M.OldGateway.LogLevel == "" {
		panic(errors.New("loglevel is not null"))
	}
	if config.M.OldGateway.LogPath == "" {
		panic(errors.New("logpath is not null "))
	}
	l, err := log.New(config.M.OldGateway.LogLevel, config.M.OldGateway.LogPath, 0)
	if err != nil {
		panic(err)
		return nil
	}
	config.Log = l
	return initialization
}

func (initialization *Initialization) bufPoolBasic() *Initialization {
	bufPool := &sync.Pool{
		New: MakeBuf,
	}
	if config.M.OldGateway.BufPool <= 1 {
		config.M.OldGateway.BufPool = 500
	}
	for i := 0; i < config.M.OldGateway.BufPool; i++ {
		bufPool.Put(bufPool.New())
	}
	initialization.bufPool = bufPool
	return initialization
}

func MakeBuf() interface{} {
	return bytes.NewBuffer(make([]byte, 0, 2048))
}

func (initialization *Initialization) initConfig(conf string) *Initialization {
	configBytes, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Error("ioutil readfile config err:%v", err)
		panic(err)
	}
	if _, err := toml.Decode(string(configBytes), &initialization.conf); err != nil {
		log.Error("toml decode err ", err)
		panic(err)
	}
	log.Debug("all config is %v", initialization.conf)
	return initialization
}

func (initialization *Initialization) serverClient() *Initialization {
	client := &sync.Pool{
		New: makeClient,
	}
	if config.M.OldGateway.BufPool <= 1 {
		config.M.OldGateway.BufPool = 100
	}
	for i := 0; i < config.M.OldGateway.BufPool; i++ {
		client.Put(client.New())
	}
	initialization.client = client
	return initialization
}

func makeClient() interface{} {
	return &http.Client{}
}

func (initialization *Initialization) do() *Initialization {
	req, err := http.NewRequest("GET", initialization.conf.ConfCenter.Addr, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")

	resp, err := makeClient().(*http.Client).Do(req)
	defer func() {
		resp.Body.Close()
	}()
	if err != nil {
		log.Error("this computer network is err ", err)
		panic(err)
	}
	log.Debug("[%v] resp.body is %v", time.Now(), resp.Body)
	body := s.Do(resp, req)
	err = json.Unmarshal(body, config.M.OldGateway)
	if err != nil {
		log.Error("json unmarshal confCenter body err", err)
		config.Log.Error("[%s] json unmarshal confCenter body err:%v", time.Now(), err)
		panic(err)
	}
	log.Debug("pull confCenter succeed,OldGateway is:%v", config.M.OldGateway)
	_, errManagerroute := config.M.Intrane.Get(config.M.OldGateway.Managerroute)
	if !errManagerroute {
		config.M.Intrane.Set(config.M.OldGateway.Managerroute, s.Configuration)
	}
	_, errServiceroute := config.M.Intrane.Get(config.M.OldGateway.Serviceroute)
	if !errServiceroute {
		config.M.Intrane.Set(config.M.OldGateway.Serviceroute, s.GetService)
	}
	return initialization
}

func (initialization *Initialization) pullServer() *Initialization {
	req, err := http.NewRequest("GET", initialization.conf.ConfCenter.SrvAddr, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")

	resp, err := makeClient().(*http.Client).Do(req)
	defer func() {
		resp.Body.Close()
	}()
	if err != nil {
		log.Error("this computer network is err ", err)
		panic(err)
	}
	log.Debug("[%v] resp.body is %v", time.Now(), resp.Body)
	body := s.Do(resp, req)
	err = json.Unmarshal(body, config.M.NewService)
	if err != nil {
		log.Error("json unmarshal confCenter body err", err)
		config.Log.Warn("[%s] json unmarshal confCenter body err:%v", time.Now(), err)
		panic(err)
	}
	log.Debug("pull confCenter succeed,NewService is:%v", config.M.NewService.Result)
	config.Log.Debug("[%v] pull confCenter succeed NewService is %v---", time.Now(), config.M.NewService.Result)
	for _, v := range config.M.NewService.Result {
		err = json.Unmarshal([]byte(v.Service), config.M.Service)
		if err != nil {
			log.Error("json unmarshalconfig.M.Service err", err)
			config.Log.Warn("[%s] json unmarshal config.M.Service err:%v", time.Now(), err)
			panic(err)
		}
		config.M.Service.ServiceName = v.ServiceName
		_, errService := config.Srv.Service.Get(v.Route)
		_, errConfig := config.Srv.Config.Get(v.Route)
		switch {
		case !errService:
			var addr []string
			for _, v := range config.M.Service.ServiceAddr {
				addr = append(addr, v)
			}
			config.Srv.Service.Set(v.Route, addr)
			fallthrough
		case !errConfig:
			log.Debug("config.Srv.Config k is", v.Route)
			config.Srv.Config.Set(v.Route, engine.Engine)
		}
		_, errBalance := config.Srv.Balance.Get(v.Route)
		if !errBalance {
			switch {
			case config.M.Service.Balance == "random":
				config.Srv.Balance.Set(v.Route, balance.NewRandom())
			case config.M.Service.Balance == "polling":
				config.Srv.Balance.Set(v.Route, balance.NewPolling())
			default:
				config.Srv.Balance.Set(v.Route, balance.NewRandom())
			}
		}
		log.Debug("pull confCenter succeed,config.Srv.Service is:%v-----------", config.Srv.Service, "v.route%v----", v.Route)
		log.Debug("finally ok ---------------------")
	}

	config.Srv.Balance.EachItem(func(item *syncmap.Item) {
		log.Debug("balance k is %v,v is %v", item.Key, item.Value)
	})

	config.Srv.Service.EachItem(func(item *syncmap.Item) {
		log.Debug("config.Srv.Service k is %v --------------v is %v", item.Key, item.Value)
	})

	config.Srv.Config.EachItem(func(item *syncmap.Item) {
		log.Debug("config.Srv.Config k is %v-----------------v is %v", item.Key, item.Value)
	})

	log.Debug("pull confCenter succeed,config.Srv.Service is:%v-----", config.Srv.Service)
	config.Log.Debug("[%v] pull confCenter succeed config.Srv.Service is %v---", time.Now(), config.Srv.Service)
	return initialization
}

func (initialization *Initialization) Reload() *Initialization {
	go syncManager()
	go syncService()
	go deleteService()
	go insertService()
	return initialization
}

func syncManager() {
	defer func() {
		if err := recover(); err != nil {
			config.Log.Error("Reload goroutine panic", err)
		}
	}()
	for i := range config.ManagerChan {
		config.Log.Debug("sync manager to gateway about OldManager--config.ManagerChan.i---%v--", i)
		switch {
		case *config.M.OldGateway != *s.Manager.OldGateway:
			fallthrough
		default:
			config.M.Intrane.Delete(config.M.OldGateway.Managerroute)
			config.M.Intrane.Delete(config.M.OldGateway.Serviceroute)
			config.M.OldGateway.Ip = s.Manager.OldGateway.Ip
			config.M.OldGateway.Port = s.Manager.OldGateway.Port
			config.M.OldGateway.TimeOut = s.Manager.OldGateway.TimeOut
			config.M.OldGateway.LogLevel = s.Manager.OldGateway.LogLevel
			config.M.OldGateway.LogPath = s.Manager.OldGateway.LogPath
			config.M.OldGateway.Modification = s.Manager.OldGateway.Modification
			config.M.OldGateway.BufPool = s.Manager.OldGateway.BufPool
			config.M.OldGateway.IntranetIp = s.Manager.OldGateway.IntranetIp
			config.M.OldGateway.IntranetPort = s.Manager.OldGateway.IntranetPort
			config.M.OldGateway.MaxHeader = s.Manager.OldGateway.MaxHeader
			config.M.OldGateway.Managerroute = s.Manager.OldGateway.Managerroute
			config.M.OldGateway.Serviceroute = s.Manager.OldGateway.Serviceroute
			config.M.Intrane.Set(config.M.OldGateway.Managerroute, s.Configuration)
			config.M.Intrane.Set(config.M.OldGateway.Serviceroute, s.GetService)
			continue
		}
	}
}

func syncService() {
	defer func() {
		if err := recover(); err != nil {
			config.Log.Error("Reload goroutine panic", err)
		}
	}()
	for i := range config.ServiceChan {
		config.Log.Debug("sync service to gateway about service------config.ServiceChan.i---%v--", i)
		for _, v1 := range config.M.NewService.Result {
			switch {
			case v1.Route != s.Service.Operations.Route:
				config.Srv.Service.Delete(v1.Route)
				config.Srv.Config.Delete(v1.Route)
				config.Srv.Balance.Delete(v1.Route)
				fallthrough
			default:
				addr := make([]string, 0, 20)
				config.M.Service.Balance = s.Service.Operations.Service.Balance
				config.M.Service.ServiceName = s.Service.Operations.ServiceName
				config.M.Service.RegisterTime = s.Service.Operations.Service.RegisterTime
				config.M.Service.AltReason = s.Service.Operations.Service.AltReason
				for _, domain := range s.Service.Operations.Service.ServiceAddr {
					addr = append(addr, domain)
				}
				config.Log.Debug("config.Srv.Service.set service router succeed %v", s.Service.Operations.Route)
				config.Srv.Service.Set(s.Service.Operations.Route, addr)

				config.Log.Debug("config.Srv.Config.Set route succeed %v", s.Service.Operations.Route)
				config.Srv.Config.Set(s.Service.Operations.Route, engine.Engine)

				switch {
				case config.M.Service.Balance == "random":
					config.Log.Debug("config.Srv.Balance.Set route succeed %v", s.Service.Operations.Route)
					config.Srv.Balance.Set(s.Service.Operations.Route, balance.NewRandom())
				case config.M.Service.Balance == "polling":
					config.Srv.Balance.Set(s.Service.Operations.Route, balance.NewPolling())
				default:
					config.Srv.Balance.Set(s.Service.Operations.Route, balance.NewRandom())
				}
				continue
			}
		}
	}
}

func deleteService() {
	defer func() {
		if err := recover(); err != nil {
			config.Log.Error("Reload goroutine panic", err)
		}
	}()
	for i := range config.DeleteChan {
		config.Log.Debug("delete service route to gateway about service----------config.DeleteChan.i---%v--", i)
		_, errService := config.Srv.Service.Get(s.DeleteService.Operations.Route)
		_, errConfig := config.Srv.Config.Get(s.DeleteService.Operations.Route)
		_, errBalance := config.Srv.Balance.Get(s.DeleteService.Operations.Route)
		switch {
		case errService:
			config.Log.Debug("config.Srv.Service.Delete service router succeed %v", s.DeleteService.Operations.Route)
			config.Srv.Service.Delete(s.DeleteService.Operations.Route)
			fallthrough
		case errConfig:
			config.Log.Debug("config.Srv.Config.Delete route succeed %v", s.DeleteService.Operations.Route)
			config.Srv.Config.Delete(s.DeleteService.Operations.Route)
			fallthrough
		case errBalance:
			config.Log.Debug("config.Srv.Balance.Delete route succeed %v", s.DeleteService.Operations.Route)
			config.Srv.Balance.Delete(s.DeleteService.Operations.Route)
			fallthrough
		default:
			continue
		}
	}
}

func insertService() {
	defer func() {
		if err := recover(); err != nil {
			config.Log.Error("Reload goroutine panic", err)
		}
	}()
	for i := range config.InsertChan {
		config.Log.Debug("insert service route to gateway about service----------config.InsertChan.i---%v--", i)
		_, errService := config.Srv.Service.Get(s.InsertService.Operations.Route)
		_, errConfig := config.Srv.Config.Get(s.InsertService.Operations.Route)
		_, errBalance := config.Srv.Balance.Get(s.InsertService.Operations.Route)
		addr := make([]string, 0, 20)
		switch {
		case !errService:
			config.Log.Debug("config.Srv.Service.set service router succeed %v", s.InsertService.Operations.Route)
			for _, v := range s.InsertService.Operations.Service.ServiceAddr {
				addr = append(addr, v)
			}
			config.Srv.Service.Set(s.InsertService.Operations.Route, addr)
			fallthrough
		case !errConfig:
			config.Log.Debug("config.Srv.Config.Set route succeed %v", s.InsertService.Operations.Route)
			config.Srv.Config.Set(s.InsertService.Operations.Route, engine.Engine)
			fallthrough
		case !errBalance:
			config.Log.Debug("config.Srv.Balance.Set route succeed %v", s.DeleteService.Operations.Route)
			switch {
			case s.InsertService.Operations.Service.Balance == "random":
				config.Srv.Balance.Set(s.InsertService.Operations.Route, balance.NewRandom())
			case s.InsertService.Operations.Service.Balance == "polling":
				config.Srv.Balance.Set(s.InsertService.Operations.Route, balance.NewPolling())
			default:
				config.Srv.Balance.Set(s.InsertService.Operations.Route, balance.NewRandom())
			}
			fallthrough
		default:
			continue
		}
	}
}
