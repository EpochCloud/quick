package engine

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"quick/balance"
	"quick/basic"
	"quick/basic/app"
	"quick/config"
)

func Engine(w http.ResponseWriter, r *http.Request) {
	service, err := validate(w, r)
	if err != nil {
		return
	}
	b, _ := config.Srv.Balance.Get(r.URL.Path)
 	domain,err :=b.(balance.Balancing).Balance(service.([]string))
	if err != nil {
		config.Log.Error("get domain err", err)
		errResult.SendErrorResponse(w, config.ErrorFailed)
		return
	}
	config.Log.Debug(domain)
	result, err := app.App(r.Method, getDomain(domain, r), r.Body)
	if err != nil {
		degradation(r.URL.Path)
		errResult.SendErrorResponse(w, config.Error)
		config.Log.Error("this service is can not conn")
		return
	}
	config.Log.Debug("service result %v,service method %v,domain %v, body %v", result, r.Method, getDomain(domain, r), r.Body)
	config.Log.Debug("engine result is header %s", result.Header)
	resp := basic.GetApp(result.Body)
	defer func() {
		result.Body.Close()
		basic.Clean(r, resp)
	}()
	getHeader(w, result)
	config.Log.Debug("engine body is %v", resp.String())
	io.WriteString(w, resp.String())
	return
}

func validate(w http.ResponseWriter, ctx *http.Request) (interface{}, error) {
	service, ok := config.Srv.Service.Get(ctx.URL.Path)
	if !ok {
		errResult.SendErrorResponse(w, config.ErrorMethodFailed)
		config.Log.Error("request route err", ctx.URL.Path)
		return nil, errors.New("request route err")
	}
	return service, nil
}

func getHeader(ctx http.ResponseWriter, w *http.Response) {
	for k, _ := range ctx.Header() {
		delete(ctx.Header(), k)
	}
	for k, v := range w.Header {
		ctx.Header()[k] = v
	}
}

func getDomain(domain string, r *http.Request) string {
	if scheme != string(domain[7]) {
		domain = fmt.Sprintf("%s%s%s", scheme, domain, r.URL.Path)
		config.Log.Debug("domian is %s", domain)
	} else {
		domain = fmt.Sprintf("%s%s", domain, r.URL.Path)
	}
	return domain
}

func degradation(p string) {
	config.Srv.Service.Delete(p)
}
