package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

const defaultServerShutdownTimeout = 30 * time.Second

func init() {
	logging.Init()
}

func main() {
	listenerCtx, stopListener := context.WithCancel(context.Background())
	defer stopListener()

	application, createContainerJobListener, cleanup := service.NewApplication(listenerCtx)
	defer cleanup()

	var listenerWG sync.WaitGroup
	if createContainerJobListener != nil {
		listenerWG.Add(1)
		go func() {
			defer listenerWG.Done()
			createContainerJobListener.Listen(listenerCtx)
		}()
	}

	httpServer := server.NewHTTPServerOnAddr(viper.GetString("server.addr"), func(router *gin.Engine) {
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

	httpErrCh := make(chan error, 1)
	go func() {
		httpErrCh <- server.RunHTTPServerInstance(httpServer)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-httpErrCh:
		if err != nil {
			panic(err)
		}
	case <-sigCh:
		shutdownTimeout := viper.GetDuration("server.shutdown-timeout")
		if shutdownTimeout <= 0 {
			shutdownTimeout = defaultServerShutdownTimeout
		}

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			cancelShutdown()
			panic(err)
		}
		cancelShutdown()

		stopListener()
		listenerWG.Wait()

		if err := <-httpErrCh; err != nil {
			panic(err)
		}
	}
}
