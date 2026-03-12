package ports

import (
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leadtek-test/q1/common"
	client "github.com/leadtek-test/q1/common/client/container"
	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/app"
	"github.com/leadtek-test/q1/container/app/command"
	"github.com/leadtek-test/q1/container/app/dto"
	"github.com/leadtek-test/q1/container/ports/contextx"
	"github.com/spf13/viper"
)

const defaultUploadFileSizeLimit = int64(20 * 1024 * 1024)

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

	r, err := H.App.Commands.CreateUser.Handle(c, command.CreateUser{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return
	}

	resp = dto.CreateUserResponse{
		UserID:   r.UserID,
		Username: r.Username,
	}
}

func (H HTTPServer) Login(c *gin.Context) {
	var (
		req  client.RegisterRequest
		resp dto.LoginUserResponse
		err  error
	)
	defer func() {
		H.Response(c, err, &resp)
	}()

	if err = c.ShouldBindJSON(&req); err != nil {
		err = errors.NewWithError(consts.ErrnoBindRequestError, err)
		return
	}

	r, err := H.App.Commands.LoginUser.Handle(c, command.LoginUser{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return
	}

	resp = dto.LoginUserResponse{
		UserID:    r.UserID,
		Username:  r.Username,
		Token:     r.Token,
		ExpiresAt: r.ExpiresAt,
	}
}

func (H HTTPServer) Upload(c *gin.Context) {
	var (
		resp dto.UploadFileResponse
		err  error
	)
	defer func() {
		H.Response(c, err, &resp)
	}()

	userID := c.GetUint(contextx.KeyUserID)
	if userID == 0 {
		err = errors.New(consts.ErrnoAuthInvalidToken)
		return
	}

	targetUserIDRaw := strings.TrimSpace(c.Param("id"))
	targetUserIDParsed, parseErr := strconv.ParseUint(targetUserIDRaw, 10, 64)
	if parseErr != nil || targetUserIDParsed == 0 {
		err = errors.NewWithMsgf(consts.ErrnoRequestValidateError, "invalid target user id: %s", targetUserIDRaw)
		return
	}
	targetUserID := uint(targetUserIDParsed)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		err = errors.NewWithError(consts.ErrnoFileRequired, err)
		return
	}

	uploadFile, err := fileHeader.Open()
	if err != nil {
		err = errors.NewWithError(consts.ErrnoFileOpenFailed, err)
		return
	}
	defer func() {
		_ = uploadFile.Close()
	}()

	maxFileSize := viper.GetInt64("file.max-size")
	if maxFileSize <= 0 {
		maxFileSize = defaultUploadFileSizeLimit
	}

	content, err := io.ReadAll(io.LimitReader(uploadFile, maxFileSize+1))
	if err != nil {
		err = errors.NewWithError(consts.ErrnoFileReadFailed, err)
		return
	}
	if int64(len(content)) > maxFileSize {
		err = errors.New(consts.ErrnoFileSizeExceeded)
		return
	}

	contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	r, err := H.App.Commands.UploadFile.Handle(c, command.UploadFile{
		UserID:      targetUserID,
		FileName:    fileHeader.Filename,
		ContentType: contentType,
		Payload:     content,
	})
	if err != nil {
		return
	}

	resp = dto.UploadFileResponse{
		ID:            r.ID,
		UserID:        r.UserID,
		FileName:      r.FileName,
		ObjectKey:     r.ObjectKey,
		ContentType:   r.ContentType,
		Size:          r.Size,
		WorkspacePath: r.WorkspacePath,
		CreatedAt:     r.CreatedAt,
	}
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
