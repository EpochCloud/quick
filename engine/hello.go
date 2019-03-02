package engine

import (
	"net/http"
	"io"
)

func Engines(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"ok")
}
