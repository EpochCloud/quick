package config

import (
	"github.com/DeanThompson/syncmap"
	"io"
	"net/http"
	"quick/log"
	"sync"
)

var (
	Buf, Client *sync.Pool
	Log         *log.Logger
	M           *Manager
	Srv         *S //这里保存service的所有的配置
	ManagerChan  chan string
	ServiceChan chan string
	DeleteChan  chan string
	InsertChan   chan string
)

type (
	Fn = func(http.ResponseWriter, *http.Request)
	S  struct {
		Service *syncmap.SyncMap
		Config  *syncmap.SyncMap
		Balance *syncmap.SyncMap
	}
	Manager struct {
		Intrane    *syncmap.SyncMap
		//NewGateway *GatewayManager
		OldGateway *GatewayManager
		NewService *ServiceOperation
		Service    *Service
		Operations *Operations
	}
	ServiceOperation struct {
		Result []Operation `json:"result"`
	}
	Operation struct {
		Id          uint64 `json:"id"`    //主id
		Route       string `json:"route"` //路由
		Service     string `json:"service"`
		ServiceName string `json:"servicename"` //服务名字
	}
	Operations struct {
		Id          uint64 `json:"id"`    //主id
		Route       string `json:"route"` //路由
		Service     *Service `json:"service"`
		ServiceName string `json:"servicename"` //服务名字
	}
	Service struct {
		ServiceAddr []string `json:"serviceaddr"` //服务地址  [ip:port]
		//RegisterName string   `json:"registername"` //谁注册的服务  这里的名字是登录的名字，不能让人填写，这里先空着，等登录注册完成之后再说补充
		RegisterTime string `json:"registertime"` //注册时间
		//AltTime      string   `json:"alttime"`      //修改时间  这里的名字是登录的名字，不能让人填写，这里先空着，等登录注册完成之后再说补充
		AltReason   string `json:"altreason"` //修改原因
		ServiceName string `json:"servicename"`
		Balance     string `json:"Balance"`
	}
	GatewayManager struct {
		Id           uint64 `json:"id"`
		Ip           string `json:"ip"`           //api服务的ip
		Port         string `json:"port"`         //api服务的端口
		TimeOut      int    `json:"timeout"`      //api设置的超时时间
		LogLevel     string `json:"loglevel"`     //日志的级别
		LogPath      string `json:"logpath"`      //日志的路径
		Modification uint64 `json:"modification"` //是否被覆盖，覆盖了是1，不覆盖是0
		BufPool      int    `json:"bufpool"`      //buf池子的容量
		IntranetIp   string `json:"intranetip"`   //内网ip
		IntranetPort string `json:"intranetport"` //内网端口
		MaxHeader    string `json:"maxheader"`    //最大请求头
		Managerroute string	`json:"managerroute"`  //配置路由
		Serviceroute string `json:"serviceroute"` //服务路由
	}
)


func NewS() *S {
	return &S{
		Service: syncmap.New(),
		Config:  syncmap.New(),
		Balance: syncmap.New(),
	}
}

func NewManagers() *Manager {
	return &Manager{
		Intrane:syncmap.New(),
		OldGateway: &GatewayManager{
		},
		NewService: &ServiceOperation{
			Result: make([]Operation, 0, 20),
		},
		Service: &Service{
			ServiceAddr: make([]string, 0, 10),
		},
		Operations:&Operations{
			Service:&Service{
				ServiceAddr:make([]string,0,10),
			},
		},
	}
}

func (m *Manager)ServeHTTP(w http.ResponseWriter, r *http.Request){
	h,err := M.Intrane.Get(r.URL.Path)
	if !err {
		io.WriteString(w, "request faild")
		return
	}
	h.(Fn)(w,r)
	return
}

func (m *S) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, err := Srv.Config.Get(r.URL.Path)
	if !err {
		io.WriteString(w, "request faild")
		return
	}
	h.(Fn)(w, r)
	return
}


