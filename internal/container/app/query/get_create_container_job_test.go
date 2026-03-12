package query

import (
	"context"
	"testing"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domainjob "github.com/leadtek-test/q1/container/domain/job"
	"github.com/sirupsen/logrus"
)

type fakeCreateContainerJobRepo struct {
	getFn func(context.Context, string, uint) (domainjob.CreateContainerJob, error)
}

func (f fakeCreateContainerJobRepo) Create(context.Context, *domainjob.CreateContainerJob) error {
	return nil
}
func (f fakeCreateContainerJobRepo) Update(context.Context, *domainjob.CreateContainerJob) error {
	return nil
}
func (f fakeCreateContainerJobRepo) GetByJobIDAndUser(ctx context.Context, jobID string, userID uint) (domainjob.CreateContainerJob, error) {
	return f.getFn(ctx, jobID, userID)
}

func TestGetCreateContainerJobHandler(t *testing.T) {
	logger := logrus.New()
	handler := NewGetCreateContainerJobHandler(
		fakeCreateContainerJobRepo{
			getFn: func(_ context.Context, jobID string, userID uint) (domainjob.CreateContainerJob, error) {
				if userID != 7 || jobID != "job-1" {
					t.Fatalf("unexpected query: userID=%d jobID=%s", userID, jobID)
				}
				return domainjob.CreateContainerJob{
					JobID:       "job-1",
					Status:      domainjob.CreateContainerJobStatusSucceeded,
					ContainerID: 21,
				}, nil
			},
		},
		logger,
	)

	result, err := handler.Handle(context.Background(), GetCreateContainerJob{
		UserID: 7,
		JobID:  "  job-1  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.JobID != "job-1" || result.Status != "succeeded" || result.ContainerID != 21 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGetCreateContainerJobHandlerErrors(t *testing.T) {
	logger := logrus.New()
	handler := NewGetCreateContainerJobHandler(
		fakeCreateContainerJobRepo{
			getFn: func(context.Context, string, uint) (domainjob.CreateContainerJob, error) {
				return domainjob.CreateContainerJob{}, commonerrors.New(consts.ErrnoContainerCreateJobNotFound)
			},
		},
		logger,
	)

	_, err := handler.Handle(context.Background(), GetCreateContainerJob{UserID: 0, JobID: "job-1"})
	if commonerrors.Errno(err) != consts.ErrnoAuthInvalidToken {
		t.Fatalf("unexpected errno: %v", err)
	}

	_, err = handler.Handle(context.Background(), GetCreateContainerJob{UserID: 1, JobID: "  "})
	if commonerrors.Errno(err) != consts.ErrnoRequestValidateError {
		t.Fatalf("unexpected errno: %v", err)
	}

	_, err = handler.Handle(context.Background(), GetCreateContainerJob{UserID: 1, JobID: "job-1"})
	if commonerrors.Errno(err) != consts.ErrnoContainerCreateJobNotFound {
		t.Fatalf("unexpected errno: %v", err)
	}
}
