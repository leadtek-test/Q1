package ports

import "github.com/gin-gonic/gin"

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
	// TODO add router register here
	//router.POST(options.BaseURL+"/customer/:customer_id/orders", handlers.PostCustomerCustomerIdOrders)
}
