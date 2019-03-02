package basic

import (
	"bytes"
	"io"
	"net/http"
	"quick/config"
)


func Clean(r *http.Request,b *bytes.Buffer){
	r.Body.Close()
	b.Reset()
	config.Buf.Put(b)
}

func GetApp(req io.ReadCloser)*bytes.Buffer{
	b := config.Buf.Get().(*bytes.Buffer)
	io.Copy(b,req)
	return b
}