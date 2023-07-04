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

type OptionServer interface {
	apply(server *http.Server)
}

type optionHttpServer func(server *http.Server)

func (fn optionHttpServer) apply(server *http.Server) {
	fn(server)
}
func WithOptionServer(fn func(server *http.Server)) OptionServer {
	return optionHttpServer(func(server *http.Server) {
		fn(server)
	})
}

func Run(opts ...OptionServer) {

	server := &http.Server{}
	for _, opt := range opts {
		opt.apply(server)
	}

	// 启动服务器
	go func() {
		log.Println("Server is start...ListenAndPort" + server.Addr)
		err := server.ListenAndServe()
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
	err := server.Shutdown(ctx)
	if err != nil {
		log.Println("Shutdown error:", err)
	}

	log.Println("Server is shut down")
}
