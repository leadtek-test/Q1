package query

import (
	"context"
	"strings"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/decorator"
	"github.com/leadtek-test/q1/common/handler/errors"
	domainjob "github.com/leadtek-test/q1/container/domain/job"
	"github.com/sirupsen/logrus"
)

type GetCreateContainerJob struct {
	UserID uint
	JobID  string
}

type GetCreateContainerJobResult struct {
	JobID        string
	Status       string
	ContainerID  uint
	ErrorMessage string
}

type GetCreateContainerJobHandler decorator.QueryHandler[GetCreateContainerJob, *GetCreateContainerJobResult]

type getCreateContainerJobHandler struct {
	repo domainjob.Repository
}

func NewGetCreateContainerJobHandler(repo domainjob.Repository, logger *logrus.Logger) GetCreateContainerJobHandler {
	if repo == nil {
		panic("get create container job's repository is nil")
	}
	return decorator.ApplyQueryDecorators(
		getCreateContainerJobHandler{repo: repo},
		logger,
	)
}

func (h getCreateContainerJobHandler) Handle(ctx context.Context, query GetCreateContainerJob) (*GetCreateContainerJobResult, error) {
	if query.UserID == 0 {
		return nil, errors.New(consts.ErrnoAuthInvalidToken)
	}
	jobID := strings.TrimSpace(query.JobID)
	if jobID == "" {
		return nil, errors.NewWithMsgf(consts.ErrnoRequestValidateError, "invalid container create job id")
	}

	job, err := h.repo.GetByJobIDAndUser(ctx, jobID, query.UserID)
	if err != nil {
		return nil, err
	}

	return &GetCreateContainerJobResult{
		JobID:        job.JobID,
		Status:       string(job.Status),
		ContainerID:  job.ContainerID,
		ErrorMessage: job.ErrorMessage,
	}, nil
}
