package adapters

import (
	"context"
	"testing"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domainjob "github.com/leadtek-test/q1/container/domain/job"
)

func TestContainerCreateJobRepositoryPostgresCRUD(t *testing.T) {
	repo := NewContainerCreateJobRepositoryPostgres(newAdaptersTestPostgres(t))
	ctx := context.Background()

	target := &domainjob.CreateContainerJob{
		JobID:   "job-1",
		UserID:  7,
		Name:    "demo",
		Image:   "busybox:latest",
		Command: []string{"echo", "ok"},
		Env:     map[string]string{"A": "1"},
		Status:  domainjob.CreateContainerJobStatusAccepted,
	}
	if err := repo.Create(ctx, target); err != nil {
		t.Fatalf("Create unexpected error: %v", err)
	}
	if target.CreatedAt.IsZero() || target.UpdatedAt.IsZero() {
		t.Fatalf("expected create to hydrate timestamps: %+v", target)
	}

	got, err := repo.GetByJobIDAndUser(ctx, "job-1", 7)
	if err != nil {
		t.Fatalf("GetByJobIDAndUser unexpected error: %v", err)
	}
	if got.JobID != "job-1" || got.Status != domainjob.CreateContainerJobStatusAccepted {
		t.Fatalf("unexpected job payload: %+v", got)
	}

	target.Status = domainjob.CreateContainerJobStatusSucceeded
	target.ContainerID = 10
	if err = repo.Update(ctx, target); err != nil {
		t.Fatalf("Update unexpected error: %v", err)
	}

	got, err = repo.GetByJobIDAndUser(ctx, "job-1", 7)
	if err != nil {
		t.Fatalf("GetByJobIDAndUser after update unexpected error: %v", err)
	}
	if got.Status != domainjob.CreateContainerJobStatusSucceeded || got.ContainerID != 10 {
		t.Fatalf("unexpected updated job payload: %+v", got)
	}
}

func TestContainerCreateJobRepositoryPostgresNotFound(t *testing.T) {
	repo := NewContainerCreateJobRepositoryPostgres(newAdaptersTestPostgres(t))

	_, err := repo.GetByJobIDAndUser(context.Background(), "missing", 1)
	if commonerrors.Errno(err) != consts.ErrnoContainerCreateJobNotFound {
		t.Fatalf("expected container create job not found errno, got err=%v", err)
	}
}
