package command

import (
	"context"
	"testing"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domainjob "github.com/leadtek-test/q1/container/domain/job"
	"github.com/sirupsen/logrus"
)

type fakeCreateContainerDispatcher struct {
	dispatchFn func(context.Context, domainjob.CreateContainerTask) (string, error)
}

func (f fakeCreateContainerDispatcher) DispatchCreateContainer(ctx context.Context, task domainjob.CreateContainerTask) (string, error) {
	return f.dispatchFn(ctx, task)
}

func (f fakeCreateContainerDispatcher) Listen(context.Context) {}

func TestCreateContainerJobHandler(t *testing.T) {
	logger := logrus.New()
	handler := NewCreateContainerJobHandler(
		fakeCreateContainerDispatcher{
			dispatchFn: func(_ context.Context, task domainjob.CreateContainerTask) (string, error) {
				if task.UserID != 3 {
					t.Fatalf("unexpected user id: %d", task.UserID)
				}
				if task.Name != "default" {
					t.Fatalf("expected default name, got %q", task.Name)
				}
				if task.Image != "busybox:latest" {
					t.Fatalf("unexpected image: %q", task.Image)
				}
				if len(task.Command) != 0 {
					t.Fatalf("expected empty command, got %+v", task.Command)
				}
				if len(task.Env) != 0 {
					t.Fatalf("expected empty env, got %+v", task.Env)
				}
				return "job-1", nil
			},
		},
		logger,
	)

	result, err := handler.Handle(context.Background(), CreateContainerJob{
		UserID: 3,
		Name:   "  ",
		Image:  "  busybox:latest  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.JobID != "job-1" {
		t.Fatalf("unexpected job id: %s", result.JobID)
	}
}

func TestCreateContainerJobHandlerErrors(t *testing.T) {
	logger := logrus.New()
	handler := NewCreateContainerJobHandler(
		fakeCreateContainerDispatcher{
			dispatchFn: func(context.Context, domainjob.CreateContainerTask) (string, error) {
				return "", commonerrors.New(consts.ErrnoDatabaseError)
			},
		},
		logger,
	)

	_, err := handler.Handle(context.Background(), CreateContainerJob{UserID: 0, Image: "img"})
	assertErrno(t, err, consts.ErrnoAuthInvalidToken)

	_, err = handler.Handle(context.Background(), CreateContainerJob{UserID: 1, Image: "  "})
	assertErrno(t, err, consts.ErrnoContainerImageRequired)

	_, err = handler.Handle(context.Background(), CreateContainerJob{UserID: 1, Image: "img"})
	assertErrno(t, err, consts.ErrnoDatabaseError)
}
