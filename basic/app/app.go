package app

import (
	"net/http"
	"quick/basic"
	"quick/config"
	"encoding/json"
	"quick/log"
	"time"
)


func Configuration(w http.ResponseWriter,r *http.Request){
	switch r.Method {
	case http.MethodPost:
		err := getConf(w,r)
		if err != nil {
			return
		}
		config.ManagerChan <- "configuration post succeed"
		return
	default:
		errResult.SendErrorResponse(w,config.ErrorMethodFailed)
		return
	}
}

func getConf(w http.ResponseWriter,r *http.Request)error{
	b := basic.GetApp(r.Body)
	defer func(){
		basic.Clean(r,b)
	}()
	err := json.Unmarshal(b.Bytes(),Manager.OldGateway)
	if err != nil {
		errResult.SendErrorResponse(w,config.ErrorJsonFailed)
		log.Error("json unmarshal confCenter manager err", err)
		config.Log.Error("[%s] json unmarshal confCenter  manager  err:%v", time.Now(), err)
		return err
	}
	config.Log.Debug("[%s] the confcenter push here manager succeed manager is %v",time.Now(),Manager.OldGateway)
	succeedResult.NormalResponse(w,succeedResult)
	return nil
}

func GetService(w http.ResponseWriter,r *http.Request){
	switch r.Method {
	case http.MethodPost:
		err := insertService(w,r)
		if err != nil {
			return
		}
		config.InsertChan <- "service post here succeed"
		return
	case http.MethodDelete:
		err := deleteService(w,r)
		if err != nil {
			return
		}
		config.DeleteChan <- "delete service here succeed"
		return
	case http.MethodPatch:
		err := getService(w,r)
		if err != nil {
			return
		}
		config.ServiceChan <- "service patch here succeed"
		return
	default:
		errResult.SendErrorResponse(w,config.ErrorMethodFailed)
		return
	}

}

func getService(w http.ResponseWriter,r *http.Request)error{
	b := basic.GetApp(r.Body)
	defer func(){
		basic.Clean(r,b)
	}()
	err := json.Unmarshal(b.Bytes(),Service.Operations)
	if err != nil {
		errResult.SendErrorResponse(w,config.ErrorJsonFailed)
		log.Error("json unmarshal confCenter service err", err)
		config.Log.Warn("[%s] json unmarshal confCenter service err:%v", time.Now(), err)
		return err
	}
	config.Log.Debug("confCenter push here Service.Operations  is %v",Service.Operations)
	succeedResult.NormalResponse(w,succeedResult)
	return nil
}

func deleteService(w http.ResponseWriter,r *http.Request)(err error){
	b := basic.GetApp(r.Body)
	defer func(){
		basic.Clean(r,b)
	}()
	err = json.Unmarshal(b.Bytes(),DeleteService.Operations)
	if err != nil {
		errResult.SendErrorResponse(w,config.ErrorJsonFailed)
		log.Error("json unmarshal confCenter service err", err)
		config.Log.Warn("[%s] json unmarshal confCenter service err:%v", time.Now(), err)
		return err
	}
	config.Log.Debug("confCenter delete here DeleteService.Operations  is %v",DeleteService.Operations)
	succeedResult.NormalResponse(w,succeedResult)
	return
}


func insertService(w http.ResponseWriter,r *http.Request)(err error){
	b := basic.GetApp(r.Body)
	defer func(){
		basic.Clean(r,b)
	}()
	err = json.Unmarshal(b.Bytes(),InsertService.Operations)
	if err != nil {
		errResult.SendErrorResponse(w,config.ErrorJsonFailed)
		log.Error("json unmarshal confCenter service err", err)
		config.Log.Warn("[%s] json unmarshal confCenter service err:%v", time.Now(), err)
		return err
	}
	config.Log.Debug("confCenter delete here DeleteService.Operations  is %v",DeleteService.Operations)
	succeedResult.NormalResponse(w,succeedResult)
	return
}

