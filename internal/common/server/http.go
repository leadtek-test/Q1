package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leadtek-test/q1/common/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func RunHTTPServer(serviceName string, wrapper func(router *gin.Engine)) {
	srv := NewHTTPServer(serviceName, wrapper)
	if err := RunHTTPServerInstance(srv); err != nil {
		panic(err)
	}
}

func NewHTTPServer(serviceName string, wrapper func(router *gin.Engine)) *http.Server {
	addr := viper.Sub(serviceName).GetString("http-addr")
	if addr == "" {
		panic("empty http address")
	}
	return NewHTTPServerOnAddr(addr, wrapper)
}

func RunHTTPServerInstance(srv *http.Server) error {
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func NewHTTPServerOnAddr(addr string, wrapper func(router *gin.Engine)) *http.Server {
	if addr == "" {
		panic("empty http address")
	}
	apiRouter := gin.New()
	setMiddlewares(apiRouter)
	wrapper(apiRouter)
	return &http.Server{
		Addr:    addr,
		Handler: apiRouter,
	}
}

func setMiddlewares(r *gin.Engine) {
	// TODO add another middleware here
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLog(logrus.NewEntry(logrus.StandardLogger())))
}
