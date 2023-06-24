package bunlk

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	HttpServer *http.Server
}

func NewServer(server *http.Server) *Server {
	return &Server{
		HttpServer: server,
	}
}

func (s *Server) Run() {
	// 启动服务器
	go func() {
		log.Println("Server is start...ListenAndPort" + s.HttpServer.Addr)
		err := s.HttpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Println("Server error:", err)
		}
	}()

	// 监听系统信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待关闭信号
	<-signalChan

	// 创建一个上下文对象，设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭服务器
	err := s.HttpServer.Shutdown(ctx)
	if err != nil {
		log.Println("Shutdown error:", err)
	}

	log.Println("Server is shut down")
}
