package bunlk

import (
	"net/http"
	"testing"
)

func TestRun(t *testing.T) {
	Run(WithOptionServer(func(server *http.Server) {
		server.Addr = ":8080"
	}))
}
