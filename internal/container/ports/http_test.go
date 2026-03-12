package ports

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/container/app"
	"github.com/leadtek-test/q1/container/app/command"
	appquery "github.com/leadtek-test/q1/container/app/query"
	"github.com/leadtek-test/q1/container/ports/contextx"
)

type envelope struct {
	Errno int             `json:"errno"`
	Data  json.RawMessage `json:"data"`
}

type fakeCreateUserHandler struct {
	fn func(context.Context, command.CreateUser) (*command.CreateUserResult, error)
}

func (f fakeCreateUserHandler) Handle(ctx context.Context, cmd command.CreateUser) (*command.CreateUserResult, error) {
	return f.fn(ctx, cmd)
}

type fakeLoginUserHandler struct {
	fn func(context.Context, command.LoginUser) (*command.LoginUserResult, error)
}

func (f fakeLoginUserHandler) Handle(ctx context.Context, cmd command.LoginUser) (*command.LoginUserResult, error) {
	return f.fn(ctx, cmd)
}

type fakeUploadFileHandler struct {
	fn func(context.Context, command.UploadFile) (*command.UploadFileResult, error)
}

func (f fakeUploadFileHandler) Handle(ctx context.Context, cmd command.UploadFile) (*command.UploadFileResult, error) {
	return f.fn(ctx, cmd)
}

type fakeCreateContainerHandler struct {
	fn func(context.Context, command.CreateContainer) (*command.CreateContainerResult, error)
}

func (f fakeCreateContainerHandler) Handle(ctx context.Context, cmd command.CreateContainer) (*command.CreateContainerResult, error) {
	return f.fn(ctx, cmd)
}

type fakeUpdateContainerStatusHandler struct {
	fn func(context.Context, command.UpdateContainerStatus) (*command.UpdateContainerStatusResult, error)
}

func (f fakeUpdateContainerStatusHandler) Handle(ctx context.Context, cmd command.UpdateContainerStatus) (*command.UpdateContainerStatusResult, error) {
	return f.fn(ctx, cmd)
}

type fakeDeleteContainerHandler struct {
	fn func(context.Context, command.DeleteContainer) (*command.DeleteContainerResult, error)
}

func (f fakeDeleteContainerHandler) Handle(ctx context.Context, cmd command.DeleteContainer) (*command.DeleteContainerResult, error) {
	return f.fn(ctx, cmd)
}

type fakeListContainersHandler struct {
	fn func(context.Context, appquery.ListContainers) (*appquery.ListContainersResult, error)
}

func (f fakeListContainersHandler) Handle(ctx context.Context, query appquery.ListContainers) (*appquery.ListContainersResult, error) {
	return f.fn(ctx, query)
}

func TestHTTPServerAuthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := HTTPServer{
		App: app.Application{
			Commands: app.Commands{
				CreateUser: fakeCreateUserHandler{
					fn: func(context.Context, command.CreateUser) (*command.CreateUserResult, error) {
						return &command.CreateUserResult{UserID: 1, Username: "u"}, nil
					},
				},
				LoginUser: fakeLoginUserHandler{
					fn: func(context.Context, command.LoginUser) (*command.LoginUserResult, error) {
						return &command.LoginUserResult{
							UserID:    1,
							Username:  "u",
							Token:     "t",
							ExpiresAt: time.Unix(100, 0),
						}, nil
					},
				},
			},
		},
	}

	// register
	ctx, w := newJSONContext(http.MethodPost, "/register", `{"username":"u","password":"123456"}`)
	server.Register(ctx)
	assertErrno(t, w, consts.ErrnoSuccess)

	// login
	ctx, w = newJSONContext(http.MethodPost, "/login", `{"username":"u","password":"123456"}`)
	server.Login(ctx)
	assertErrno(t, w, consts.ErrnoSuccess)
}

func TestHTTPServerUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := HTTPServer{
		App: app.Application{
			Commands: app.Commands{
				UploadFile: fakeUploadFileHandler{
					fn: func(_ context.Context, cmd command.UploadFile) (*command.UploadFileResult, error) {
						if cmd.UserID != 2 {
							t.Fatalf("unexpected target user: %d", cmd.UserID)
						}
						return &command.UploadFileResult{
							ID:        8,
							UserID:    cmd.UserID,
							FileName:  "a.txt",
							ObjectKey: "users/2/x_a.txt",
						}, nil
					},
				},
			},
		},
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "a.txt")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err = part.Write([]byte("abc")); err != nil {
		t.Fatalf("write part failed: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/user/2/files", &body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx.Params = gin.Params{{Key: "id", Value: "2"}}
	ctx.Set(contextx.KeyUserID, uint(9))

	server.Upload(ctx)
	assertErrno(t, w, consts.ErrnoSuccess)
}

func TestHTTPServerContainerEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := HTTPServer{
		App: app.Application{
			Commands: app.Commands{
				CreateContainer: fakeCreateContainerHandler{
					fn: func(_ context.Context, cmd command.CreateContainer) (*command.CreateContainerResult, error) {
						return &command.CreateContainerResult{
							ID:        1,
							UserID:    cmd.UserID,
							Name:      cmd.Name,
							Image:     cmd.Image,
							Status:    "created",
							CreatedAt: time.Unix(1, 0),
							UpdatedAt: time.Unix(2, 0),
						}, nil
					},
				},
				UpdateContainerStatus: fakeUpdateContainerStatusHandler{
					fn: func(_ context.Context, cmd command.UpdateContainerStatus) (*command.UpdateContainerStatusResult, error) {
						return &command.UpdateContainerStatusResult{
							ID:      cmd.ContainerID,
							UserID:  cmd.UserID,
							Status:  "running",
							Name:    "n",
							Image:   "img",
							Command: []string{},
							Env:     map[string]string{},
						}, nil
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
						return &appquery.ListContainersResult{Containers: []appquery.ContainerItem{{ID: 1}}}, nil
					},
				},
			},
		},
	}

	// create
	ctx, w := newJSONContext(http.MethodPost, "/containers", `{"name":"n","image":"img"}`)
	ctx.Set(contextx.KeyUserID, uint(3))
	server.CreateContainer(ctx)
	assertErrno(t, w, consts.ErrnoSuccess)

	// list
	ctx, w = newJSONContext(http.MethodGet, "/containers", ``)
	ctx.Set(contextx.KeyUserID, uint(3))
	server.ListContainers(ctx)
	assertErrno(t, w, consts.ErrnoSuccess)

	// update status
	ctx, w = newJSONContext(http.MethodPut, "/containers/1/status", `{"action":"start"}`)
	ctx.Set(contextx.KeyUserID, uint(3))
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
	server.UpdateContainerStatus(ctx)
	assertErrno(t, w, consts.ErrnoSuccess)

	// delete
	ctx, w = newJSONContext(http.MethodDelete, "/containers/1", ``)
	ctx.Set(contextx.KeyUserID, uint(3))
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
	server.DeleteContainer(ctx)
	assertErrno(t, w, consts.ErrnoSuccess)
}

func TestParseContainerID(t *testing.T) {
	id, err := parseContainerID("42")
	if err != nil || id != 42 {
		t.Fatalf("unexpected parse result id=%d err=%v", id, err)
	}

	_, err = parseContainerID("bad")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func newJSONContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	return ctx, w
}

func assertErrno(t *testing.T, w *httptest.ResponseRecorder, errno int) {
	t.Helper()
	var resp envelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v, body=%s", err, w.Body.String())
	}
	if resp.Errno != errno {
		t.Fatalf("unexpected errno: got %d want %d body=%s", resp.Errno, errno, w.Body.String())
	}
}
