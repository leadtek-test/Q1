package common

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leadtek-test/q1/common/handler/errors"
)

type BaseResponse struct{}

type response struct {
	Errno   int    `json:"errno"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func (base *BaseResponse) Response(c *gin.Context, err error, data interface{}) {
	if err != nil {
		base.error(c, err)
	} else {
		base.success(c, data)
	}
}

func (base *BaseResponse) success(c *gin.Context, data interface{}) {
	errno, errmsg := errors.Output(nil)
	r := response{
		Errno:   errno,
		Message: errmsg,
		Data:    data,
	}
	resp, _ := json.Marshal(r)
	c.Set("response", string(resp))
	c.JSON(http.StatusOK, r)
}

func (base *BaseResponse) error(c *gin.Context, err error) {
	errno, errmsg, statusCode := errors.OutputWithStatus(err)
	r := response{
		Errno:   errno,
		Message: errmsg,
		Data:    nil,
	}
	resp, _ := json.Marshal(r)
	c.Set("response", string(resp))
	c.JSON(statusCode, r)
}
