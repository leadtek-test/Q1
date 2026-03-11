package ports

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServerOptions struct {
	BaseURL      string
	Middlewares  []gin.HandlerFunc
	ErrorHandler func(*gin.Context, error, int)
}

func RegisterHandlersWithOption(router gin.IRouter, handlers ServerInterface, options ServerOptions) {
	errorHandler := options.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}
	if len(options.Middlewares) > 0 {
		for _, m := range options.Middlewares {
			router.Use(m)
		}
	}

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1Group := router.Group(options.BaseURL + "/v1")
	{
		auth := v1Group.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		protected := v1Group.Group("")
		protected.Use()
		{
			protected.POST("/files", handlers.Upload)

			protected.POST("/containers", handlers.CreateContainer)
			protected.GET("/containers", handlers.ListContainers)
			protected.PUT("/containers/status", handlers.UpdateContainerStatus)
			protected.DELETE("/containers/:id", handlers.DeleteContainer)
		}
	}
}
