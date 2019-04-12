package initialize

import (
	"bytes"
	"encoding/json"
	"errors"
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
	config.ManagerChan = make(chan interface{}, 10)
}

func Initialize(conf string) {
	newInitialization := NewInitialization().
		initConfig(conf).
		do().
		logInitialize().
		bufPoolBasic().
		serverClient().
		pullServer().
		Reload()
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

//start pull the config from ConfCenter
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
		switch i.(type) {
		//sync quick config
		case *config.GatewayManager:
			syncGetwayConfig(i.(*config.GatewayManager))
			continue
		case *config.Operations:
			insertOperations(i.(*config.Operations))
			continue
		case *syncmap.SyncMap:
			m := i.(*syncmap.SyncMap)
			deletOperation, errDelet := m.Get(http.MethodDelete)
			patchOperation, errPatch := m.Get(http.MethodPatch)
			switch {
			case errDelet:
				deleteOperatons(deletOperation.(*config.Operations))
				continue
			case errPatch:
				patchOperations(patchOperation.(*config.Operations))
				continue
			}
			continue
		}
	}
}

func syncGetwayConfig(i *config.GatewayManager) {
	switch {
	case *config.M.OldGateway != *i:
		fallthrough
	default:
		config.M.Intrane.Delete(config.M.OldGateway.Managerroute)
		config.M.Intrane.Delete(config.M.OldGateway.Serviceroute)
		config.M.OldGateway.Ip = i.Ip
		config.M.OldGateway.Port = i.Port
		config.M.OldGateway.TimeOut = i.TimeOut
		config.M.OldGateway.LogLevel = i.LogLevel
		config.M.OldGateway.LogPath = i.LogPath
		config.M.OldGateway.Modification = i.Modification
		config.M.OldGateway.BufPool = i.BufPool
		config.M.OldGateway.IntranetIp = i.IntranetIp
		config.M.OldGateway.IntranetPort = i.IntranetPort
		config.M.OldGateway.MaxHeader = i.MaxHeader
		config.M.OldGateway.Managerroute = i.Managerroute
		config.M.OldGateway.Serviceroute = i.Serviceroute
		//just change key
		config.M.Intrane.Set(config.M.OldGateway.Managerroute, s.Configuration)
		config.M.Intrane.Set(config.M.OldGateway.Serviceroute, s.GetService)
		//todo change  other like log and reload server
	}
}

func patchOperations(i *config.Operations) {
	for _, v1 := range config.M.NewService.Result {
		switch {
		case v1.Route != i.Route:
			config.Srv.Service.Delete(v1.Route)
			config.Srv.Config.Delete(v1.Route)
			config.Srv.Balance.Delete(v1.Route)
			fallthrough
		default:
			addr := make([]string, 0, 20)
			config.M.Service.Balance = i.Service.Balance
			config.M.Service.ServiceName = i.ServiceName
			config.M.Service.RegisterTime = i.Service.RegisterTime
			config.M.Service.AltReason = i.Service.AltReason
			for _, domain := range i.Service.ServiceAddr {
				addr = append(addr, domain)
			}
			config.Log.Debug("config.Srv.Service.set service router succeed %v", s.Service.Operations.Route)
			config.Srv.Service.Set(i.Route, addr)

			config.Log.Debug("config.Srv.Config.Set route succeed %v", s.Service.Operations.Route)
			config.Srv.Config.Set(i.Route, engine.Engine)

			switch {
			case config.M.Service.Balance == "random":
				config.Log.Debug("config.Srv.Balance.Set route succeed %v", s.Service.Operations.Route)
				config.Srv.Balance.Set(i.Route, balance.NewRandom())
			case config.M.Service.Balance == "polling":
				config.Srv.Balance.Set(i.Route, balance.NewPolling())
			default:
				config.Srv.Balance.Set(i.Route, balance.NewRandom())
			}
			continue
		}
	}
}

func deleteOperatons(i *config.Operations) {
	_, errService := config.Srv.Service.Get(i.Route)
	_, errConfig := config.Srv.Config.Get(i.Route)
	_, errBalance := config.Srv.Balance.Get(i.Route)
	switch {
	case errService:
		config.Log.Debug("config.Srv.Service.Delete service router succeed %v", s.DeleteService.Operations.Route)
		config.Srv.Service.Delete(i.Route)
		fallthrough
	case errConfig:
		config.Log.Debug("config.Srv.Config.Delete route succeed %v", s.DeleteService.Operations.Route)
		config.Srv.Config.Delete(i.Route)
		fallthrough
	case errBalance:
		config.Log.Debug("config.Srv.Balance.Delete route succeed %v", s.DeleteService.Operations.Route)
		config.Srv.Balance.Delete(i.Route)
		fallthrough
	default:
	}
}

func insertOperations(i *config.Operations) {
	config.Log.Debug("insert service route to gateway about service----------config.InsertChan.i---%v--", i)
	_, errService := config.Srv.Service.Get(i.Route)
	_, errConfig := config.Srv.Config.Get(i.Route)
	_, errBalance := config.Srv.Balance.Get(i.Route)
	addr := make([]string, 0, 20)
	switch {
	case !errService:
		config.Log.Debug("config.Srv.Service.set service router succeed %v", s.InsertService.Operations.Route)
		for _, v := range i.Service.ServiceAddr {
			addr = append(addr, v)
		}
		config.Srv.Service.Set(i.Route, addr)
		fallthrough
	case !errConfig:
		config.Log.Debug("config.Srv.Config.Set route succeed %v", s.InsertService.Operations.Route)
		config.Srv.Config.Set(i.Route, engine.Engine)
		fallthrough
	case !errBalance:
		config.Log.Debug("config.Srv.Balance.Set route succeed %v", s.DeleteService.Operations.Route)
		switch {
		case i.Service.Balance == "random":
			config.Srv.Balance.Set(i.Route, balance.NewRandom())
		case i.Service.Balance == "polling":
			config.Srv.Balance.Set(i.Route, balance.NewPolling())
		default:
			config.Srv.Balance.Set(i.Route, balance.NewRandom())
		}
		fallthrough
	default:
		//continue
	}
}