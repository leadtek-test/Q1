package command

import (
	"context"
	stderrors "errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domainauth "github.com/leadtek-test/q1/container/domain/auth"
	domaincontainer "github.com/leadtek-test/q1/container/domain/container"
	domainfile "github.com/leadtek-test/q1/container/domain/file"
	domainuser "github.com/leadtek-test/q1/container/domain/user"
	"github.com/sirupsen/logrus"
)

type fakeUserRepo struct {
	createFn        func(context.Context, *domainuser.User) error
	getByUsernameFn func(context.Context, string) (domainuser.User, error)
	getByIDFn       func(context.Context, uint) (domainuser.User, error)
}

func (f fakeUserRepo) Create(ctx context.Context, u *domainuser.User) error {
	if f.createFn != nil {
		return f.createFn(ctx, u)
	}
	return nil
}

func (f fakeUserRepo) GetByUsername(ctx context.Context, username string) (domainuser.User, error) {
	if f.getByUsernameFn != nil {
		return f.getByUsernameFn(ctx, username)
	}
	return domainuser.User{}, commonerrors.New(consts.ErrnoUserNotFound)
}

func (f fakeUserRepo) GetByID(ctx context.Context, id uint) (domainuser.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return domainuser.User{}, commonerrors.New(consts.ErrnoUserNotFound)
}

type fakeHasher struct {
	hashFn    func(string) string
	compareFn func(string, string) bool
}

func (f fakeHasher) Hash(raw string) string {
	if f.hashFn != nil {
		return f.hashFn(raw)
	}
	return "hashed:" + raw
}

func (f fakeHasher) Compare(raw, encoded string) bool {
	if f.compareFn != nil {
		return f.compareFn(raw, encoded)
	}
	return f.Hash(raw) == encoded
}

type fakeTokenManager struct {
	generateFn func(uint, string) (string, time.Time, error)
	parseFn    func(string) (domainauth.Claims, error)
}

func (f fakeTokenManager) Generate(userID uint, username string) (string, time.Time, error) {
	if f.generateFn != nil {
		return f.generateFn(userID, username)
	}
	return "token", time.Now().Add(time.Minute), nil
}

func (f fakeTokenManager) Parse(token string) (domainauth.Claims, error) {
	if f.parseFn != nil {
		return f.parseFn(token)
	}
	return domainauth.Claims{}, nil
}

type fakeFileRepo struct {
	createFn func(context.Context, *domainfile.File) error
}

func (f fakeFileRepo) Create(ctx context.Context, file *domainfile.File) error {
	if f.createFn != nil {
		return f.createFn(ctx, file)
	}
	return nil
}

type fakeObjectStorage struct {
	uploadFn func(context.Context, string, int64, string) error
}

func (f fakeObjectStorage) Upload(ctx context.Context, key string, _ io.Reader, size int64, contentType string) error {
	if f.uploadFn != nil {
		return f.uploadFn(ctx, key, size, contentType)
	}
	return nil
}

type fakeWorkspace struct {
	ensureFn func(uint) (string, error)
	saveFn   func(uint, string, []byte) (string, error)
}

func (f fakeWorkspace) EnsureUserDir(userID uint) (string, error) {
	if f.ensureFn != nil {
		return f.ensureFn(userID)
	}
	return "/tmp/test", nil
}

func (f fakeWorkspace) Save(userID uint, fileName string, data []byte) (string, error) {
	if f.saveFn != nil {
		return f.saveFn(userID, fileName, data)
	}
	return "/tmp/test/" + fileName, nil
}

type fakeContainerRepo struct {
	createFn      func(context.Context, *domaincontainer.Container) error
	getByIDUserFn func(context.Context, uint, uint) (domaincontainer.Container, error)
	updateFn      func(context.Context, *domaincontainer.Container) error
	deleteFn      func(context.Context, uint, uint) error
	listByUserFn  func(context.Context, uint) ([]domaincontainer.Container, error)
}

func (f fakeContainerRepo) Create(ctx context.Context, c *domaincontainer.Container) error {
	if f.createFn != nil {
		return f.createFn(ctx, c)
	}
	return nil
}

func (f fakeContainerRepo) GetByIDAndUser(ctx context.Context, id, userID uint) (domaincontainer.Container, error) {
	if f.getByIDUserFn != nil {
		return f.getByIDUserFn(ctx, id, userID)
	}
	return domaincontainer.Container{}, commonerrors.New(consts.ErrnoContainerNotFound)
}

func (f fakeContainerRepo) Update(ctx context.Context, c *domaincontainer.Container) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, c)
	}
	return nil
}

func (f fakeContainerRepo) Delete(ctx context.Context, id, userID uint) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id, userID)
	}
	return nil
}

func (f fakeContainerRepo) ListByUser(ctx context.Context, userID uint) ([]domaincontainer.Container, error) {
	if f.listByUserFn != nil {
		return f.listByUserFn(ctx, userID)
	}
	return nil, nil
}

type fakeContainerRuntime struct {
	createFn func(context.Context, uint, domaincontainer.CreateSpec, string) (string, error)
	startFn  func(context.Context, string) error
	stopFn   func(context.Context, string) error
	deleteFn func(context.Context, string) error
}

func (f fakeContainerRuntime) Create(ctx context.Context, userID uint, spec domaincontainer.CreateSpec, workspacePath string) (string, error) {
	if f.createFn != nil {
		return f.createFn(ctx, userID, spec, workspacePath)
	}
	return "runtime-id", nil
}

func (f fakeContainerRuntime) Start(ctx context.Context, runtimeID string) error {
	if f.startFn != nil {
		return f.startFn(ctx, runtimeID)
	}
	return nil
}

func (f fakeContainerRuntime) Stop(ctx context.Context, runtimeID string) error {
	if f.stopFn != nil {
		return f.stopFn(ctx, runtimeID)
	}
	return nil
}

func (f fakeContainerRuntime) Delete(ctx context.Context, runtimeID string) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, runtimeID)
	}
	return nil
}

func TestCreateUserHandler(t *testing.T) {
	logger := logrus.New()
	handler := NewCreateUserHandler(
		fakeUserRepo{
			getByUsernameFn: func(context.Context, string) (domainuser.User, error) {
				return domainuser.User{}, commonerrors.New(consts.ErrnoUserNotFound)
			},
			createFn: func(_ context.Context, u *domainuser.User) error {
				u.ID = 10
				return nil
			},
		},
		fakeHasher{},
		logger,
	)

	result, err := handler.Handle(context.Background(), CreateUser{
		Username: " alice ",
		Password: " secret1 ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != 10 || result.Username != "alice" {
		t.Fatalf("unexpected result: %+v", result)
	}

	_, err = handler.Handle(context.Background(), CreateUser{Username: "", Password: "secret1"})
	assertErrno(t, err, consts.ErrnoUserUsernameRequired)
}

func TestLoginUserHandler(t *testing.T) {
	logger := logrus.New()
	handler := NewLoginUserHandler(
		fakeUserRepo{
			getByUsernameFn: func(context.Context, string) (domainuser.User, error) {
				return domainuser.User{ID: 1, Username: "alice", PasswordHashed: "x"}, nil
			},
		},
		fakeHasher{compareFn: func(string, string) bool { return true }},
		fakeTokenManager{
			generateFn: func(userID uint, username string) (string, time.Time, error) {
				return "tk", time.Unix(100, 0), nil
			},
		},
		logger,
	)

	result, err := handler.Handle(context.Background(), LoginUser{Username: "alice", Password: "123456"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token != "tk" {
		t.Fatalf("unexpected token: %s", result.Token)
	}

	_, err = handler.Handle(context.Background(), LoginUser{Username: "", Password: "123456"})
	assertErrno(t, err, consts.ErrnoUserUsernameRequired)
}

func TestUploadFileHandler(t *testing.T) {
	logger := logrus.New()
	handler := NewUploadFileHandler(
		fakeFileRepo{
			createFn: func(_ context.Context, f *domainfile.File) error {
				f.ID = 88
				return nil
			},
		},
		fakeObjectStorage{},
		fakeWorkspace{
			saveFn: func(_ uint, name string, _ []byte) (string, error) {
				if !strings.HasSuffix(name, "_a.txt") {
					t.Fatalf("unexpected store file name: %s", name)
				}
				return "/tmp/" + name, nil
			},
		},
		1024,
		logger,
	)

	res, err := handler.Handle(context.Background(), UploadFile{
		UserID:   3,
		FileName: "a.txt",
		Payload:  []byte("abc"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 88 || !strings.HasPrefix(res.ObjectKey, "users/3/") {
		t.Fatalf("unexpected result: %+v", res)
	}

	_, err = handler.Handle(context.Background(), UploadFile{
		UserID:   3,
		FileName: "a.txt",
		Payload:  make([]byte, 1025),
	})
	assertErrno(t, err, consts.ErrnoFileSizeExceeded)
}

func TestUpdateContainerStatusHandler(t *testing.T) {
	logger := logrus.New()
	startCalled := 0
	stopCalled := 0
	updateCalled := 0

	handler := NewUpdateContainerStatusHandler(
		fakeContainerRepo{
			getByIDUserFn: func(_ context.Context, id, userID uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{
					ID:        id,
					UserID:    userID,
					RuntimeID: "r1",
					Status:    domaincontainer.StatusStopped,
				}, nil
			},
			updateFn: func(_ context.Context, c *domaincontainer.Container) error {
				updateCalled++
				if c.Status != domaincontainer.StatusRunning {
					t.Fatalf("expected running, got %s", c.Status)
				}
				return nil
			},
		},
		fakeContainerRuntime{
			startFn: func(context.Context, string) error {
				startCalled++
				return nil
			},
			stopFn: func(context.Context, string) error {
				stopCalled++
				return nil
			},
		},
		logger,
	)

	_, err := handler.Handle(context.Background(), UpdateContainerStatus{
		UserID:      1,
		ContainerID: 1,
		Action:      "start",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if startCalled != 1 || stopCalled != 0 || updateCalled != 1 {
		t.Fatalf("unexpected counters start=%d stop=%d update=%d", startCalled, stopCalled, updateCalled)
	}

	_, err = handler.Handle(context.Background(), UpdateContainerStatus{
		UserID:      1,
		ContainerID: 1,
		Action:      "invalid",
	})
	assertErrno(t, err, consts.ErrnoContainerInvalidStatusAction)
}

func TestDeleteContainerHandler(t *testing.T) {
	logger := logrus.New()
	deleteCalled := 0

	handler := NewDeleteContainerHandler(
		fakeContainerRepo{
			getByIDUserFn: func(context.Context, uint, uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{RuntimeID: "r1"}, nil
			},
			deleteFn: func(context.Context, uint, uint) error {
				deleteCalled++
				return nil
			},
		},
		fakeContainerRuntime{
			deleteFn: func(context.Context, string) error { return nil },
		},
		logger,
	)

	res, err := handler.Handle(context.Background(), DeleteContainer{
		UserID:      1,
		ContainerID: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Deleted || deleteCalled != 1 {
		t.Fatalf("unexpected result: %+v deleteCalled=%d", res, deleteCalled)
	}

	handlerFail := NewDeleteContainerHandler(
		fakeContainerRepo{
			getByIDUserFn: func(context.Context, uint, uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{RuntimeID: "r1"}, nil
			},
		},
		fakeContainerRuntime{
			deleteFn: func(context.Context, string) error { return stderrors.New("boom") },
		},
		logger,
	)
	_, err = handlerFail.Handle(context.Background(), DeleteContainer{UserID: 1, ContainerID: 2})
	assertErrno(t, err, consts.ErrnoContainerRuntimeDeleteFail)
}

func assertErrno(t *testing.T, err error, code int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got := commonerrors.Errno(err); got != code {
		t.Fatalf("unexpected errno: got %d, want %d, err=%v", got, code, err)
	}
}
