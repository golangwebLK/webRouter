package bunlk

import (
	"net/http"
	"testing"
)

func TestRun(t *testing.T) {
	s := &Server{
		HttpServer: &http.Server{
			Addr:    ":8080",
			Handler: nil, // 这里应该设置你的请求处理程序
		},
	}
	s.Run()
}
