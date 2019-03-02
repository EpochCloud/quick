package config

import (
	"encoding/json"
	"io"
	"net/http"
)

var (

	ErrorFailed = ErrorResult{
		HttpCode: 401,
		Error: Err{
			Error:     "remote service err",
			ErrorCode: "001",
		},
	}

	ErrorJsonFailed = ErrorResult{
		HttpCode: 500,
		Error: Err{
			Error:     "json marshal or unmarshal  err",
			ErrorCode: "002",
		},
	}

	ErrorMethodFailed = ErrorResult{
		HttpCode: 409,
		Error: Err{
			Error:     "request err",
			ErrorCode: "009",
		},
	}

	Error = ErrorResult{
		HttpCode: 500,
		Error: Err{
			Error:     "service err",
			ErrorCode: "019",
		},
	}
)

type Err struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code"`
}

type ErrorResult struct {
	Error    Err
	HttpCode int
}

func NewErrorResult() *ErrorResult {
	return &ErrorResult{}
}

func (r ErrorResult) SendErrorResponse(w http.ResponseWriter, errResponse ErrorResult) {
	w.WriteHeader(errResponse.HttpCode)
	errMessage, _ := json.Marshal(&errResponse.Error)
	io.WriteString(w,string(errMessage))
}

type NormalResult struct {
	Resp string
	Code int
}

func NewResult() *NormalResult {
	return &NormalResult{
		Resp: "ok",
		Code: 200,
	}
}

func (r *NormalResult) Response(w http.ResponseWriter) {
	w.WriteHeader(r.Code)
	io.WriteString(w, r.Resp)
}

func (r *NormalResult) NormalResponse(w http.ResponseWriter, result *NormalResult) {
	w.WriteHeader(result.Code)
	io.WriteString(w, result.Resp)
}
