package app

import (
	"net/http"
	"time"
	"quick/config"
	"io"
)

func Proto(method,domain string,body io.ReadCloser)(error,*http.Request){
	req,err := http.NewRequest(method,domain,body)
	if err != nil {
		config.Log.Error("[%v] http.request host err",time.Now(),err)
		return err,nil
	}
	req.Header.Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	req.Header.Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	//req.Header.Set("content-type", "application/json")
	req.Header.Add("Content-Type","application/json;charset=UTF-8")
	return nil,req
}


func HttpDo(method,domain string,body io.ReadCloser)(result *http.Response,err error){
	err,req := Proto(method,domain,body)
	if err != nil {
		return
	}
	c := config.Client.Get()
	result,err = c.(*http.Client).Do(req)
	if err != nil {
		config.Log.Error("[%v] http.request host err",time.Now(),err)
		return
	}
	config.Log.Debug("result is %s",result)
	defer func(c interface{}){
		config.Client.Put(c)
		if err := recover();err != nil {
			config.Log.Error("[%v] this App goroutine err",time.Now(),err)
			return
		}
	}(c)
	if result.StatusCode == 200 {
		config.Log.Debug("[%v] successfull about app method %s,host %s",time.Now(),method,domain)
		return
	}
	config.Log.Error("[%v] the resp code  is ",time.Now(),result.StatusCode)
	return
}

func App(method,domain string,b io.ReadCloser) (*http.Response,error){
	switch  {
	case method == http.MethodGet:
		result,err := HttpDo(method,domain,nil)
		if err != nil {
			return nil,err
		}
		return result,nil
	default:
		result,err := HttpDo(method,domain,b)
		if err != nil {
			return nil,err
		}
		return result,nil
	}
}
