package ports

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServerOptions struct {
	BaseURL              string
	Middlewares          []gin.HandlerFunc
	ProtectedMiddlewares []gin.HandlerFunc
	ErrorHandler         func(*gin.Context, error, int)
}

func RegisterHandlersWithOption(router gin.IRouter, handlers ServerInterface, options ServerOptions) {
	errorHandler := options.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}
	useMiddlewares(router, options.Middlewares)

	router.GET("/healthz", func(c *gin.Context) {
		response := gin.H{"status": "ok"}
		c.Set("response", response)
		c.JSON(http.StatusOK, response)
	})

	v1Group := router.Group(options.BaseURL + "/v1")
	{
		auth := v1Group.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		protected := v1Group.Group("")
		useMiddlewares(protected, options.ProtectedMiddlewares)
		{
			protected.POST("/user/:id/files", handlers.Upload)

			protected.POST("/containers", handlers.CreateContainer)
			protected.GET("/containers", handlers.ListContainers)
			protected.PUT("/containers/:id/status", handlers.UpdateContainerStatus)
			protected.DELETE("/containers/:id", handlers.DeleteContainer)
		}
	}
}

type middlewareRegistrar interface {
	Use(...gin.HandlerFunc) gin.IRoutes
}

func useMiddlewares(target middlewareRegistrar, middlewares []gin.HandlerFunc) {
	if len(middlewares) == 0 {
		return
	}
	target.Use(middlewares...)
}
