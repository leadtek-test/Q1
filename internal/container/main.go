package main

import (
	"context"

	"github.com/gin-gonic/gin"
	_ "github.com/leadtek-test/q1/common/config"
	"github.com/leadtek-test/q1/common/logging"
	"github.com/leadtek-test/q1/common/server"
	"github.com/leadtek-test/q1/container/adapters"
	"github.com/leadtek-test/q1/container/ports"
	"github.com/leadtek-test/q1/container/ports/middleware"
	"github.com/leadtek-test/q1/container/service"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, createContainerJobListener, cleanup := service.NewApplication(ctx)
	defer cleanup()
	if createContainerJobListener != nil {
		go createContainerJobListener.Listen(ctx)
	}

	server.RunHTTPServerOnAddr(viper.GetString("server-addr"), func(router *gin.Engine) {
		ports.RegisterHandlersWithOption(router, ports.HTTPServer{
			App: application,
		}, ports.ServerOptions{
			BaseURL: "/api",
			ProtectedMiddlewares: []gin.HandlerFunc{
				middleware.NewAuthMiddleware(adapters.NewTokenManagerRepositoryJWT(
					viper.GetString("security.jwt-secret"),
					viper.GetDuration("security.jwt-expire-time"),
				)).VerifyToken(),
			},
		})
	})
}
