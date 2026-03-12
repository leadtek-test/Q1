package server

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestSetMiddlewares(t *testing.T) {
	r := gin.New()
	setMiddlewares(r)
	if len(r.Handlers) == 0 {
		t.Fatalf("middlewares should be registered")
	}
}

func TestRunHTTPServerPanicOnEmptyAddr(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	viper.Set("svc.http-addr", "")
	RunHTTPServer("svc", func(*gin.Engine) {})
}
