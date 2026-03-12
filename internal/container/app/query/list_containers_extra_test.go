package query

import (
	"context"
	"testing"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domaincontainer "github.com/leadtek-test/q1/container/domain/container"
	"github.com/sirupsen/logrus"
)

func TestNewListContainersHandlerPanics(t *testing.T) {
	logger := logrus.New()

	assertPanic(t, func() { NewListContainersHandler(nil, fakeRuntime{}, logger) })
	assertPanic(t, func() { NewListContainersHandler(fakeContainerRepo{}, nil, logger) })
}

func TestListContainersHandlerRepoError(t *testing.T) {
	logger := logrus.New()
	handler := NewListContainersHandler(
		fakeContainerRepo{
			listFn: func(context.Context, uint) ([]domaincontainer.Container, error) {
				return nil, commonerrors.New(consts.ErrnoDatabaseError)
			},
		},
		fakeRuntime{},
		logger,
	)

	_, err := handler.Handle(context.Background(), ListContainers{UserID: 1})
	if commonerrors.Errno(err) != consts.ErrnoDatabaseError {
		t.Fatalf("expected db error errno, got err=%v", err)
	}
}

func assertPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}
