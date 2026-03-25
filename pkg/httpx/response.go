package httpx

import "net/http"

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(data interface{}) (int, Response) {
	return http.StatusOK, Response{Code: 0, Message: "OK", Data: data}
}

func Fail(code int, msg string) (int, Response) {
	return http.StatusOK, Response{Code: code, Message: msg}
}
