package adapters

import (
	"context"
	"testing"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domaincontainer "github.com/leadtek-test/q1/container/domain/container"
	"github.com/leadtek-test/q1/container/infrastructure/persistent"
	"gorm.io/gorm"
)

func TestNormalizeAndMarshalHelpers(t *testing.T) {
	command, err := normalizeAndMarshalCommand(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if command != "[]" {
		t.Fatalf("unexpected command json: %s", command)
	}

	env, err := normalizeAndMarshalEnv(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env != "{}" {
		t.Fatalf("unexpected env json: %s", env)
	}
}

func TestToDomainContainer(t *testing.T) {
	model := persistent.ContainerModel{
		UserID:    1,
		Name:      "n",
		Image:     "img",
		Command:   `["a","b"]`,
		Env:       `{"k":"v"}`,
		RuntimeID: "rid",
		Status:    string(domaincontainer.StatusRunning),
	}
	model.ID = 9
	model.CreatedAt = time.Unix(1, 0)
	model.UpdatedAt = time.Unix(2, 0)

	result, err := toDomainContainer(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 9 || len(result.Command) != 2 || result.Env["k"] != "v" {
		t.Fatalf("unexpected result: %+v", result)
	}

	_, err = toDomainContainer(persistent.ContainerModel{Command: "{bad json}"})
	if got := commonerrors.Errno(err); got != consts.ErrnoDatabaseError {
		t.Fatalf("unexpected errno: %d", got)
	}

	_, err = toDomainContainer(persistent.ContainerModel{Env: "{bad json}"})
	if got := commonerrors.Errno(err); got != consts.ErrnoDatabaseError {
		t.Fatalf("unexpected env errno: %d", got)
	}
}

func TestContainerRepositoryPostgresCRUD(t *testing.T) {
	repo := NewContainerRepositoryPostgres(newAdaptersTestPostgres(t))
	ctx := context.Background()

	data := &domaincontainer.Container{
		UserID:    3,
		Name:      "demo",
		Image:     "busybox:latest",
		Command:   []string{"sleep", "3"},
		Env:       map[string]string{"A": "1"},
		RuntimeID: "runtime-id",
		Status:    domaincontainer.StatusCreated,
	}
	if err := repo.Create(ctx, data); err != nil {
		t.Fatalf("Create unexpected error: %v", err)
	}
	if data.ID == 0 || data.CreatedAt.IsZero() || data.UpdatedAt.IsZero() {
		t.Fatalf("Create should hydrate domain id/timestamps, got %+v", data)
	}

	items, err := repo.ListByUser(ctx, data.UserID)
	if err != nil {
		t.Fatalf("ListByUser unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID == 0 {
		t.Fatalf("unexpected list result after create: %+v", items)
	}
	data.ID = items[0].ID

	got, err := repo.GetByIDAndUser(ctx, data.ID, data.UserID)
	if err != nil {
		t.Fatalf("GetByIDAndUser unexpected error: %v", err)
	}
	if got.Name != "demo" || got.RuntimeID != "runtime-id" {
		t.Fatalf("unexpected container: %+v", got)
	}

	data.Name = "demo-v2"
	data.Status = domaincontainer.StatusRunning
	data.Command = nil
	data.Env = nil
	if err = repo.Update(ctx, data); err != nil {
		t.Fatalf("Update unexpected error: %v", err)
	}

	got, err = repo.GetByIDAndUser(ctx, data.ID, data.UserID)
	if err != nil {
		t.Fatalf("GetByIDAndUser after update unexpected error: %v", err)
	}
	if got.Name != "demo-v2" || got.Status != domaincontainer.StatusRunning {
		t.Fatalf("unexpected updated container: %+v", got)
	}
	if got.Command == nil || got.Env == nil {
		t.Fatalf("expected normalized command/env, got command=%v env=%v", got.Command, got.Env)
	}

	items, err = repo.ListByUser(ctx, data.UserID)
	if err != nil {
		t.Fatalf("ListByUser unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != data.ID {
		t.Fatalf("unexpected list result: %+v", items)
	}

	if err = repo.Delete(ctx, data.ID, data.UserID); err != nil {
		t.Fatalf("Delete unexpected error: %v", err)
	}

	_, err = repo.GetByIDAndUser(ctx, data.ID, data.UserID)
	if commonerrors.Errno(err) != consts.ErrnoContainerNotFound {
		t.Fatalf("expected not found after delete, got err=%v", err)
	}
}

func TestContainerRepositoryPostgresListByUserBadJSON(t *testing.T) {
	pg := newAdaptersTestPostgres(t)
	ctx := context.Background()
	repo := NewContainerRepositoryPostgres(pg)

	raw := &persistent.ContainerModel{
		UserID:    11,
		Name:      "demo",
		Image:     "busybox:latest",
		Command:   "{bad json}",
		Env:       "{}",
		RuntimeID: "rid",
		Status:    string(domaincontainer.StatusCreated),
	}
	if err := pg.StartTransaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Create(raw).Error
	}); err != nil {
		t.Fatalf("seed invalid container json failed: %v", err)
	}

	_, err := repo.ListByUser(ctx, 11)
	if commonerrors.Errno(err) != consts.ErrnoDatabaseError {
		t.Fatalf("expected database error for bad json, got err=%v", err)
	}
}
