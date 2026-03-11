package main

import (
	"context"

	"github.com/gin-gonic/gin"
	_ "github.com/leadtek-test/q1/common/config"
	"github.com/leadtek-test/q1/common/logging"
	"github.com/leadtek-test/q1/common/server"
	"github.com/leadtek-test/q1/container/ports"
	"github.com/leadtek-test/q1/container/service"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, cleanup := service.NewApplication(ctx)
	defer cleanup()

	server.RunHTTPServerOnAddr(viper.GetString("server-addr"), func(router *gin.Engine) {
		ports.RegisterHandlersWithOption(router, ports.HTTPServer{
			App: application,
		}, ports.ServerOptions{
			BaseURL: "/api",
		})
	})
}
