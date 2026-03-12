package command

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domaincontainer "github.com/leadtek-test/q1/container/domain/container"
	"github.com/sirupsen/logrus"
)

func TestContainerActionMutexDeleteThenStart(t *testing.T) {
	logger := logrus.New()
	locker := NewInMemoryContainerActionLocker()

	var deleted int32
	var startCalled int32
	deleteEntered := make(chan struct{}, 1)
	releaseDelete := make(chan struct{})
	deleteDone := make(chan error, 1)
	startDone := make(chan error, 1)

	repo := fakeContainerRepo{
		getByIDUserFn: func(_ context.Context, id, userID uint) (domaincontainer.Container, error) {
			if atomic.LoadInt32(&deleted) == 1 {
				return domaincontainer.Container{}, commonerrors.New(consts.ErrnoContainerNotFound)
			}
			return domaincontainer.Container{
				ID:        id,
				UserID:    userID,
				RuntimeID: "runtime-1",
				Status:    domaincontainer.StatusStopped,
			}, nil
		},
		deleteFn: func(context.Context, uint, uint) error {
			atomic.StoreInt32(&deleted, 1)
			return nil
		},
	}
	runtime := fakeContainerRuntime{
		startFn: func(context.Context, string) error {
			atomic.AddInt32(&startCalled, 1)
			return nil
		},
		deleteFn: func(context.Context, string) error {
			deleteEntered <- struct{}{}
			<-releaseDelete
			return nil
		},
	}

	deleteHandler := NewDeleteContainerHandler(repo, runtime, logger, locker)
	startHandler := NewUpdateContainerStatusHandler(repo, runtime, logger, locker)

	go func() {
		_, err := deleteHandler.Handle(context.Background(), DeleteContainer{
			UserID:      1,
			ContainerID: 9,
		})
		deleteDone <- err
	}()

	select {
	case <-deleteEntered:
	case <-time.After(time.Second):
		t.Fatal("delete operation did not enter runtime delete in time")
	}

	go func() {
		_, err := startHandler.Handle(context.Background(), UpdateContainerStatus{
			UserID:      1,
			ContainerID: 9,
			Action:      "start",
		})
		startDone <- err
	}()

	select {
	case err := <-startDone:
		t.Fatalf("start operation should wait for delete lock, got early err=%v", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(releaseDelete)

	if err := <-deleteDone; err != nil {
		t.Fatalf("delete operation failed: %v", err)
	}

	err := <-startDone
	if err == nil {
		t.Fatal("expected start operation to fail after container deletion")
	}
	if commonerrors.Errno(err) != consts.ErrnoContainerNotFound {
		t.Fatalf("unexpected errno: %d err=%v", commonerrors.Errno(err), err)
	}
	if !strings.Contains(err.Error(), "deleted by another request") {
		t.Fatalf("unexpected error message: %v", err)
	}
	if atomic.LoadInt32(&startCalled) != 0 {
		t.Fatalf("start runtime should not be called after deletion, got=%d", atomic.LoadInt32(&startCalled))
	}
}
