package adapters

import (
	"context"
	"encoding/json"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/handler/errors"
	domainjob "github.com/leadtek-test/q1/container/domain/job"
	"github.com/leadtek-test/q1/container/infrastructure/persistent"
)

type ContainerCreateJobRepositoryPostgres struct {
	db *persistent.Postgres
}

func NewContainerCreateJobRepositoryPostgres(db *persistent.Postgres) *ContainerCreateJobRepositoryPostgres {
	return &ContainerCreateJobRepositoryPostgres{db: db}
}

func (c ContainerCreateJobRepositoryPostgres) Create(ctx context.Context, job *domainjob.CreateContainerJob) error {
	commandData, err := normalizeAndMarshalCommand(job.Command)
	if err != nil {
		return err
	}
	envData, err := normalizeAndMarshalEnv(job.Env)
	if err != nil {
		return err
	}

	model := &persistent.ContainerCreateJobModel{
		JobID:        job.JobID,
		UserID:       job.UserID,
		Name:         job.Name,
		Image:        job.Image,
		Command:      commandData,
		Env:          envData,
		Status:       string(job.Status),
		ErrorMessage: job.ErrorMessage,
		ContainerID:  job.ContainerID,
	}
	if err = c.db.CreateContainerCreateJob(ctx, nil, model); err != nil {
		return err
	}

	job.CreatedAt = model.CreatedAt
	job.UpdatedAt = model.UpdatedAt
	return nil
}

func (c ContainerCreateJobRepositoryPostgres) GetByJobIDAndUser(ctx context.Context, jobID string, userID uint) (domainjob.CreateContainerJob, error) {
	model, err := c.db.GetContainerCreateJobByJobIDAndUser(ctx, jobID, userID)
	if err != nil {
		return domainjob.CreateContainerJob{}, err
	}
	return toDomainCreateContainerJob(*model)
}

func (c ContainerCreateJobRepositoryPostgres) Update(ctx context.Context, job *domainjob.CreateContainerJob) error {
	model := &persistent.ContainerCreateJobModel{
		JobID:        job.JobID,
		UserID:       job.UserID,
		Status:       string(job.Status),
		ErrorMessage: job.ErrorMessage,
		ContainerID:  job.ContainerID,
	}
	return c.db.UpdateContainerCreateJob(ctx, nil, model)
}

func toDomainCreateContainerJob(model persistent.ContainerCreateJobModel) (domainjob.CreateContainerJob, error) {
	commandData := make([]string, 0)
	if model.Command != "" {
		if err := json.Unmarshal([]byte(model.Command), &commandData); err != nil {
			return domainjob.CreateContainerJob{}, errors.NewWithError(consts.ErrnoDatabaseError, err)
		}
	}

	envData := map[string]string{}
	if model.Env != "" {
		if err := json.Unmarshal([]byte(model.Env), &envData); err != nil {
			return domainjob.CreateContainerJob{}, errors.NewWithError(consts.ErrnoDatabaseError, err)
		}
	}

	return domainjob.CreateContainerJob{
		JobID:        model.JobID,
		UserID:       model.UserID,
		Name:         model.Name,
		Image:        model.Image,
		Command:      commandData,
		Env:          envData,
		Status:       domainjob.CreateContainerJobStatus(model.Status),
		ErrorMessage: model.ErrorMessage,
		ContainerID:  model.ContainerID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}
