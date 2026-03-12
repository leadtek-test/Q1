package ports

import (
	"github.com/gin-gonic/gin"
	"github.com/leadtek-test/q1/common"
	client "github.com/leadtek-test/q1/common/client/container"
	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/app"
	"github.com/leadtek-test/q1/container/app/dto"
)

type HTTPServer struct {
	common.BaseResponse
	App app.Application
}

func (H HTTPServer) Register(c *gin.Context) {
	var (
		req  client.RegisterRequest
		resp dto.CreateUserResponse
		err  error
	)
	defer func() {
		H.Response(c, err, &resp)
	}()

	if err = c.ShouldBindJSON(&req); err != nil {
		err = errors.NewWithError(consts.ErrnoBindRequestError, err)
		return
	}

	r, err := H.App.Commands.CreateUser.Handle(c, req)
	if err != nil {
		return
	}

	resp = dto.CreateUserResponse{
		UserID:   r.UserID,
		Username: r.Username,
	}
}

func (H HTTPServer) Login(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) Upload(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) CreateContainer(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) ListContainers(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) UpdateContainerStatus(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) DeleteContainer(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
