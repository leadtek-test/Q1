package command

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domaincontainer "github.com/leadtek-test/q1/container/domain/container"
	domainfile "github.com/leadtek-test/q1/container/domain/file"
	domainuser "github.com/leadtek-test/q1/container/domain/user"
	"github.com/sirupsen/logrus"
)

func TestNewHandlersPanicsOnNilDependency(t *testing.T) {
	logger := logrus.New()
	assertPanic(t, func() { NewCreateUserHandler(nil, fakeHasher{}, logger) })
	assertPanic(t, func() { NewCreateUserHandler(fakeUserRepo{}, nil, logger) })

	assertPanic(t, func() { NewLoginUserHandler(nil, fakeHasher{}, fakeTokenManager{}, logger) })
	assertPanic(t, func() { NewLoginUserHandler(fakeUserRepo{}, nil, fakeTokenManager{}, logger) })
	assertPanic(t, func() { NewLoginUserHandler(fakeUserRepo{}, fakeHasher{}, nil, logger) })

	assertPanic(t, func() { NewUploadFileHandler(nil, fakeObjectStorage{}, fakeWorkspace{}, 1, logger) })
	assertPanic(t, func() { NewUploadFileHandler(fakeFileRepo{}, nil, fakeWorkspace{}, 1, logger) })
	assertPanic(t, func() { NewUploadFileHandler(fakeFileRepo{}, fakeObjectStorage{}, nil, 1, logger) })
	assertPanic(t, func() { NewUploadFileHandler(fakeFileRepo{}, fakeObjectStorage{}, fakeWorkspace{}, 0, logger) })

	assertPanic(t, func() { NewCreateContainerJobHandler(nil, logger) })

	assertPanic(t, func() { NewUpdateContainerStatusHandler(nil, fakeContainerRuntime{}, logger) })
	assertPanic(t, func() { NewUpdateContainerStatusHandler(fakeContainerRepo{}, nil, logger) })
	assertPanic(t, func() { NewDeleteContainerHandler(nil, fakeContainerRuntime{}, logger) })
	assertPanic(t, func() { NewDeleteContainerHandler(fakeContainerRepo{}, nil, logger) })
}

func TestCreateUserHandlerErrorBranches(t *testing.T) {
	logger := logrus.New()

	handler := NewCreateUserHandler(
		fakeUserRepo{
			getByUsernameFn: func(context.Context, string) (domainuser.User, error) {
				return domainuser.User{}, nil
			},
		},
		fakeHasher{},
		logger,
	)
	_, err := handler.Handle(context.Background(), CreateUser{Username: "alice", Password: "123456"})
	assertErrno(t, err, consts.ErrnoUserAlreadyExists)

	handler = NewCreateUserHandler(
		fakeUserRepo{
			getByUsernameFn: func(context.Context, string) (domainuser.User, error) {
				return domainuser.User{}, commonerrors.New(consts.ErrnoDatabaseError)
			},
		},
		fakeHasher{},
		logger,
	)
	_, err = handler.Handle(context.Background(), CreateUser{Username: "alice", Password: "123456"})
	assertErrno(t, err, consts.ErrnoDatabaseError)

	handler = NewCreateUserHandler(
		fakeUserRepo{
			getByUsernameFn: func(context.Context, string) (domainuser.User, error) {
				return domainuser.User{}, commonerrors.New(consts.ErrnoUserNotFound)
			},
			createFn: func(context.Context, *domainuser.User) error { return commonerrors.New(consts.ErrnoDatabaseError) },
		},
		fakeHasher{},
		logger,
	)
	_, err = handler.Handle(context.Background(), CreateUser{Username: "alice", Password: "123456"})
	assertErrno(t, err, consts.ErrnoDatabaseError)

	_, err = handler.Handle(context.Background(), CreateUser{Username: "alice", Password: ""})
	assertErrno(t, err, consts.ErrnoUserPasswordRequired)
	_, err = handler.Handle(context.Background(), CreateUser{Username: "alice", Password: "123"})
	assertErrno(t, err, consts.ErrnoUserPasswordTooShort)
}

func TestLoginUserHandlerErrorBranches(t *testing.T) {
	logger := logrus.New()
	baseUser := domainuser.User{ID: 1, Username: "alice", PasswordHashed: "hashed"}

	handler := NewLoginUserHandler(
		fakeUserRepo{
			getByUsernameFn: func(context.Context, string) (domainuser.User, error) {
				return baseUser, commonerrors.New(consts.ErrnoDatabaseError)
			},
		},
		fakeHasher{},
		fakeTokenManager{},
		logger,
	)
	_, err := handler.Handle(context.Background(), LoginUser{Username: "alice", Password: "123456"})
	assertErrno(t, err, consts.ErrnoDatabaseError)

	handler = NewLoginUserHandler(
		fakeUserRepo{
			getByUsernameFn: func(context.Context, string) (domainuser.User, error) {
				return baseUser, nil
			},
		},
		fakeHasher{compareFn: func(string, string) bool { return false }},
		fakeTokenManager{},
		logger,
	)
	_, err = handler.Handle(context.Background(), LoginUser{Username: "alice", Password: "123456"})
	assertErrno(t, err, consts.ErrnoUserPasswordNotMatch)

	handler = NewLoginUserHandler(
		fakeUserRepo{
			getByUsernameFn: func(context.Context, string) (domainuser.User, error) {
				return baseUser, nil
			},
		},
		fakeHasher{compareFn: func(string, string) bool { return true }},
		fakeTokenManager{
			generateFn: func(uint, string) (string, time.Time, error) {
				return "", time.Time{}, commonerrors.New(consts.ErrnoDatabaseError)
			},
		},
		logger,
	)
	_, err = handler.Handle(context.Background(), LoginUser{Username: "alice", Password: "123456"})
	assertErrno(t, err, consts.ErrnoDatabaseError)

	handler = NewLoginUserHandler(
		fakeUserRepo{
			getByUsernameFn: func(context.Context, string) (domainuser.User, error) {
				return domainuser.User{}, commonerrors.New(consts.ErrnoUserNotFound)
			},
		},
		fakeHasher{},
		fakeTokenManager{},
		logger,
	)
	_, err = handler.Handle(context.Background(), LoginUser{Username: "alice", Password: "123456"})
	assertErrno(t, err, consts.ErrnoUserNotFound)

	_, err = handler.Handle(context.Background(), LoginUser{Username: "alice", Password: ""})
	assertErrno(t, err, consts.ErrnoUserPasswordRequired)
	_, err = handler.Handle(context.Background(), LoginUser{Username: "alice", Password: "123"})
	assertErrno(t, err, consts.ErrnoUserPasswordTooShort)
}

func TestUploadFileHandlerErrorBranches(t *testing.T) {
	logger := logrus.New()

	handler := NewUploadFileHandler(
		fakeFileRepo{},
		fakeObjectStorage{
			uploadFn: func(context.Context, string, int64, string) error {
				return stderrors.New("upload fail")
			},
		},
		fakeWorkspace{},
		1024,
		logger,
	)

	_, err := handler.Handle(context.Background(), UploadFile{UserID: 1, FileName: "a.txt", Payload: []byte("x")})
	assertErrno(t, err, consts.ErrnoFileUploadFailed)

	handler = NewUploadFileHandler(
		fakeFileRepo{},
		fakeObjectStorage{},
		fakeWorkspace{
			saveFn: func(uint, string, []byte) (string, error) {
				return "", stderrors.New("save fail")
			},
		},
		1024,
		logger,
	)
	_, err = handler.Handle(context.Background(), UploadFile{UserID: 1, FileName: "a.txt", Payload: []byte("x")})
	assertErrno(t, err, consts.ErrnoFileWorkspaceSaveFail)

	handler = NewUploadFileHandler(
		fakeFileRepo{
			createFn: func(context.Context, *domainfile.File) error {
				return commonerrors.New(consts.ErrnoDatabaseError)
			},
		},
		fakeObjectStorage{},
		fakeWorkspace{},
		1024,
		logger,
	)
	_, err = handler.Handle(context.Background(), UploadFile{UserID: 1, FileName: "a.txt", Payload: []byte("x")})
	assertErrno(t, err, consts.ErrnoDatabaseError)

	_, err = handler.Handle(context.Background(), UploadFile{UserID: 0, FileName: "a.txt", Payload: []byte("x")})
	assertErrno(t, err, consts.ErrnoAuthInvalidToken)
	_, err = handler.Handle(context.Background(), UploadFile{UserID: 1, FileName: "", Payload: []byte("x")})
	assertErrno(t, err, consts.ErrnoFileNameRequired)
}

func TestUpdateContainerStatusHandlerAdditionalBranches(t *testing.T) {
	logger := logrus.New()

	handler := NewUpdateContainerStatusHandler(
		fakeContainerRepo{
			getByIDUserFn: func(context.Context, uint, uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{ID: 1, UserID: 1, RuntimeID: "r", Status: domaincontainer.StatusRunning}, nil
			},
		},
		fakeContainerRuntime{},
		logger,
	)
	res, err := handler.Handle(context.Background(), UpdateContainerStatus{UserID: 1, ContainerID: 1, Action: "running"})
	if err != nil || res.Status != string(domaincontainer.StatusRunning) {
		t.Fatalf("expected idempotent running, got res=%+v err=%v", res, err)
	}

	handler = NewUpdateContainerStatusHandler(
		fakeContainerRepo{
			getByIDUserFn: func(context.Context, uint, uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{ID: 1, UserID: 1, RuntimeID: "r", Status: domaincontainer.StatusStopped}, nil
			},
		},
		fakeContainerRuntime{},
		logger,
	)
	res, err = handler.Handle(context.Background(), UpdateContainerStatus{UserID: 1, ContainerID: 1, Action: "stopped"})
	if err != nil || res.Status != string(domaincontainer.StatusStopped) {
		t.Fatalf("expected idempotent stopped, got res=%+v err=%v", res, err)
	}

	handler = NewUpdateContainerStatusHandler(
		fakeContainerRepo{
			getByIDUserFn: func(context.Context, uint, uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{ID: 1, UserID: 1, RuntimeID: "r", Status: domaincontainer.StatusStopped}, nil
			},
			updateFn: func(context.Context, *domaincontainer.Container) error {
				return commonerrors.New(consts.ErrnoDatabaseError)
			},
		},
		fakeContainerRuntime{
			startFn: func(context.Context, string) error { return nil },
		},
		logger,
	)
	_, err = handler.Handle(context.Background(), UpdateContainerStatus{UserID: 1, ContainerID: 1, Action: "start"})
	assertErrno(t, err, consts.ErrnoDatabaseError)

	handler = NewUpdateContainerStatusHandler(
		fakeContainerRepo{
			getByIDUserFn: func(context.Context, uint, uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{ID: 1, UserID: 1, RuntimeID: "r", Status: domaincontainer.StatusStopped}, nil
			},
		},
		fakeContainerRuntime{
			startFn: func(context.Context, string) error { return stderrors.New("start fail") },
		},
		logger,
	)
	_, err = handler.Handle(context.Background(), UpdateContainerStatus{UserID: 1, ContainerID: 1, Action: "start"})
	assertErrno(t, err, consts.ErrnoContainerRuntimeStartFail)

	handler = NewUpdateContainerStatusHandler(
		fakeContainerRepo{
			getByIDUserFn: func(context.Context, uint, uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{ID: 1, UserID: 1, RuntimeID: "r", Status: domaincontainer.StatusRunning}, nil
			},
		},
		fakeContainerRuntime{
			stopFn: func(context.Context, string) error { return stderrors.New("stop fail") },
		},
		logger,
	)
	_, err = handler.Handle(context.Background(), UpdateContainerStatus{UserID: 1, ContainerID: 1, Action: "stop"})
	assertErrno(t, err, consts.ErrnoContainerRuntimeStopFail)

	_, err = handler.Handle(context.Background(), UpdateContainerStatus{UserID: 0, ContainerID: 1, Action: "start"})
	assertErrno(t, err, consts.ErrnoAuthInvalidToken)
	_, err = handler.Handle(context.Background(), UpdateContainerStatus{UserID: 1, ContainerID: 0, Action: "start"})
	assertErrno(t, err, consts.ErrnoRequestValidateError)
}

func TestDeleteContainerHandlerAdditionalBranches(t *testing.T) {
	logger := logrus.New()
	handler := NewDeleteContainerHandler(
		fakeContainerRepo{
			getByIDUserFn: func(context.Context, uint, uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{}, commonerrors.New(consts.ErrnoContainerNotFound)
			},
		},
		fakeContainerRuntime{},
		logger,
	)
	_, err := handler.Handle(context.Background(), DeleteContainer{UserID: 1, ContainerID: 1})
	assertErrno(t, err, consts.ErrnoContainerNotFound)

	handler = NewDeleteContainerHandler(
		fakeContainerRepo{
			getByIDUserFn: func(context.Context, uint, uint) (domaincontainer.Container, error) {
				return domaincontainer.Container{RuntimeID: "r"}, nil
			},
			deleteFn: func(context.Context, uint, uint) error { return commonerrors.New(consts.ErrnoDatabaseError) },
		},
		fakeContainerRuntime{},
		logger,
	)
	_, err = handler.Handle(context.Background(), DeleteContainer{UserID: 1, ContainerID: 1})
	assertErrno(t, err, consts.ErrnoDatabaseError)

	_, err = handler.Handle(context.Background(), DeleteContainer{UserID: 0, ContainerID: 1})
	assertErrno(t, err, consts.ErrnoAuthInvalidToken)
	_, err = handler.Handle(context.Background(), DeleteContainer{UserID: 1, ContainerID: 0})
	assertErrno(t, err, consts.ErrnoRequestValidateError)
}

func assertPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic, got nil")
		}
	}()
	fn()
}
