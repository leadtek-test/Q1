package query

import (
	"context"
	"testing"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domaincontainer "github.com/leadtek-test/q1/container/domain/container"
	"github.com/sirupsen/logrus"
)

type fakeContainerRepo struct {
	listFn func(context.Context, uint) ([]domaincontainer.Container, error)
}

func (f fakeContainerRepo) Create(context.Context, *domaincontainer.Container) error { return nil }
func (f fakeContainerRepo) GetByIDAndUser(context.Context, uint, uint) (domaincontainer.Container, error) {
	return domaincontainer.Container{}, nil
}
func (f fakeContainerRepo) Update(context.Context, *domaincontainer.Container) error { return nil }
func (f fakeContainerRepo) Delete(context.Context, uint, uint) error                 { return nil }
func (f fakeContainerRepo) ListByUser(ctx context.Context, userID uint) ([]domaincontainer.Container, error) {
	return f.listFn(ctx, userID)
}

type fakeRuntime struct{}

func (f fakeRuntime) Create(context.Context, uint, domaincontainer.CreateSpec, string) (string, error) {
	return "", nil
}
func (f fakeRuntime) Start(context.Context, string) error  { return nil }
func (f fakeRuntime) Stop(context.Context, string) error   { return nil }
func (f fakeRuntime) Delete(context.Context, string) error { return nil }

func TestListContainersHandler(t *testing.T) {
	logger := logrus.New()

	handler := NewListContainersHandler(
		fakeContainerRepo{
			listFn: func(context.Context, uint) ([]domaincontainer.Container, error) {
				return []domaincontainer.Container{
					{
						ID:        1,
						UserID:    9,
						Name:      "n",
						Image:     "img",
						Status:    domaincontainer.StatusRunning,
						CreatedAt: time.Unix(10, 0),
						UpdatedAt: time.Unix(20, 0),
					},
				}, nil
			},
		},
		fakeRuntime{},
		logger,
	)

	result, err := handler.Handle(context.Background(), ListContainers{UserID: 9})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Containers) != 1 || result.Containers[0].Status != string(domaincontainer.StatusRunning) {
		t.Fatalf("unexpected result: %+v", result)
	}

	_, err = handler.Handle(context.Background(), ListContainers{UserID: 0})
	if got := commonerrors.Errno(err); got != consts.ErrnoAuthInvalidToken {
		t.Fatalf("unexpected errno: %d", got)
	}
}
