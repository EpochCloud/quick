package app

import (
	"net/http"
	"io/ioutil"
	"quick/log"
	//"ApiGetway/config"
)


func Do(resp *http.Response,r *http.Request)[]byte{
	body,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	log.Debug("body---",string(body))
	return body
}

