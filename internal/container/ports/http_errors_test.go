package ports

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/container/app"
	"github.com/leadtek-test/q1/container/app/command"
	appquery "github.com/leadtek-test/q1/container/app/query"
	"github.com/leadtek-test/q1/container/ports/contextx"
	"github.com/spf13/viper"
)

func TestHTTPServerRegisterAndLoginBindErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := HTTPServer{App: app.Application{}}

	ctx, w := newJSONContext(http.MethodPost, "/register", "{")
	server.Register(ctx)
	assertNonSuccess(t, w)

	ctx, w = newJSONContext(http.MethodPost, "/login", "{")
	server.Login(ctx)
	assertNonSuccess(t, w)
}

func TestHTTPServerUploadErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := HTTPServer{
		App: app.Application{
			Commands: app.Commands{
				UploadFile: fakeUploadFileHandler{
					fn: func(_ context.Context, cmd command.UploadFile) (*command.UploadFileResult, error) {
						return &command.UploadFileResult{
							ID:            1,
							UserID:        cmd.UserID,
							FileName:      cmd.FileName,
							ObjectKey:     "users/1/x",
							ContentType:   cmd.ContentType,
							Size:          int64(len(cmd.Payload)),
							WorkspacePath: "/tmp/x",
						}, nil
					},
				},
			},
		},
	}

	// missing user id in context
	ctx, w := newJSONContext(http.MethodPost, "/api/v1/user/1/files", "")
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
	server.Upload(ctx)
	assertNonSuccess(t, w)

	// invalid target user id
	ctx, w = newJSONContext(http.MethodPost, "/api/v1/user/bad/files", "")
	ctx.Set(contextx.KeyUserID, uint(9))
	ctx.Params = gin.Params{{Key: "id", Value: "bad"}}
	server.Upload(ctx)
	assertNonSuccess(t, w)

	// missing form file
	ctx, w = newJSONContext(http.MethodPost, "/api/v1/user/1/files", "")
	ctx.Set(contextx.KeyUserID, uint(9))
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
	server.Upload(ctx)
	assertNonSuccess(t, w)

	// file too large
	viper.Set("file.max-size", int64(1))
	defer viper.Set("file.max-size", int64(0))
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "a.txt")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err = part.Write([]byte("ab")); err != nil {
		t.Fatalf("write part failed: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}
	w = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/user/1/files", &body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx.Set(contextx.KeyUserID, uint(9))
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
	server.Upload(ctx)
	assertNonSuccess(t, w)
}

func TestHTTPServerContainerErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := HTTPServer{
		App: app.Application{
			Commands: app.Commands{
				CreateContainer: fakeCreateContainerHandler{
					fn: func(context.Context, command.CreateContainer) (*command.CreateContainerResult, error) {
						return &command.CreateContainerResult{}, nil
					},
				},
				UpdateContainerStatus: fakeUpdateContainerStatusHandler{
					fn: func(context.Context, command.UpdateContainerStatus) (*command.UpdateContainerStatusResult, error) {
						return &command.UpdateContainerStatusResult{}, nil
					},
				},
				DeleteContainer: fakeDeleteContainerHandler{
					fn: func(context.Context, command.DeleteContainer) (*command.DeleteContainerResult, error) {
						return &command.DeleteContainerResult{Deleted: true}, nil
					},
				},
			},
			Queries: app.Queries{
				ListContainers: fakeListContainersHandler{
					fn: func(context.Context, appquery.ListContainers) (*appquery.ListContainersResult, error) {
						return &appquery.ListContainersResult{}, nil
					},
				},
			},
		},
	}

	ctx, w := newJSONContext(http.MethodPost, "/containers", `{"name":"n","image":"img"}`)
	server.CreateContainer(ctx)
	assertNonSuccess(t, w)

	ctx, w = newJSONContext(http.MethodPost, "/containers", `{`)
	ctx.Set(contextx.KeyUserID, uint(1))
	server.CreateContainer(ctx)
	assertNonSuccess(t, w)

	ctx, w = newJSONContext(http.MethodGet, "/containers", "")
	server.ListContainers(ctx)
	assertNonSuccess(t, w)

	ctx, w = newJSONContext(http.MethodPut, "/containers/bad/status", `{"action":"start"}`)
	ctx.Set(contextx.KeyUserID, uint(1))
	ctx.Params = gin.Params{{Key: "id", Value: "bad"}}
	server.UpdateContainerStatus(ctx)
	assertNonSuccess(t, w)

	ctx, w = newJSONContext(http.MethodPut, "/containers/1/status", `{`)
	ctx.Set(contextx.KeyUserID, uint(1))
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
	server.UpdateContainerStatus(ctx)
	assertNonSuccess(t, w)

	ctx, w = newJSONContext(http.MethodDelete, "/containers/1", "")
	server.DeleteContainer(ctx)
	assertNonSuccess(t, w)

	ctx, w = newJSONContext(http.MethodDelete, "/containers/bad", "")
	ctx.Set(contextx.KeyUserID, uint(1))
	ctx.Params = gin.Params{{Key: "id", Value: "bad"}}
	server.DeleteContainer(ctx)
	assertNonSuccess(t, w)
}

func assertNonSuccess(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	var resp envelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v, body=%s", err, w.Body.String())
	}
	if resp.Errno == consts.ErrnoSuccess {
		t.Fatalf("expected non-success errno, body=%s", w.Body.String())
	}
}
