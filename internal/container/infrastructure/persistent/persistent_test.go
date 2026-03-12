package persistent

import (
	"context"
	stderrors "errors"
	"fmt"
	"testing"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/infrastructure/persistent/builder"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestPostgres(t *testing.T) *Postgres {
	t.Helper()

	dsn := fmt.Sprintf("file:container_persistent_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err = db.AutoMigrate(&UserModel{}, &ContainerModel{}, &FileModel{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return NewPostgresWithDB(db)
}

func TestPostgresTransactionHelpers(t *testing.T) {
	pg := newTestPostgres(t)

	if got := pg.UseTransaction(nil); got == nil {
		t.Fatalf("UseTransaction(nil) should return base db")
	}

	tx := pg.db.Begin()
	if got := pg.UseTransaction(tx); got != tx {
		t.Fatalf("UseTransaction should return provided tx")
	}
	_ = tx.Rollback()

	err := pg.StartTransaction(func(tx *gorm.DB) error {
		return tx.Create(&UserModel{Username: "tx_user", Password: "pwd"}).Error
	})
	if err != nil {
		t.Fatalf("StartTransaction unexpected error: %v", err)
	}

	wantErr := stderrors.New("rollback-me")
	err = pg.StartTransaction(func(*gorm.DB) error { return wantErr })
	if err == nil || err.Error() != wantErr.Error() {
		t.Fatalf("expected transaction error %v, got %v", wantErr, err)
	}
}

func TestPostgresUserCRUD(t *testing.T) {
	pg := newTestPostgres(t)
	ctx := context.Background()

	create := &UserModel{Username: "alice", Password: "hashed"}
	if err := pg.CreateUser(ctx, nil, create); err != nil {
		t.Fatalf("CreateUser unexpected error: %v", err)
	}
	if create.ID == 0 || create.CreatedAt.IsZero() || create.UpdatedAt.IsZero() {
		t.Fatalf("CreateUser should hydrate id/timestamps, got %+v", create)
	}

	got, err := pg.GetUser(ctx, builder.NewUser().Usernames("alice"))
	if err != nil {
		t.Fatalf("GetUser unexpected error: %v", err)
	}
	if got.Username != "alice" || got.Password != "hashed" {
		t.Fatalf("unexpected user payload: %+v", got)
	}

	list, err := pg.BatchGetUser(ctx, builder.NewUser().Usernames("alice"))
	if err != nil {
		t.Fatalf("BatchGetUser unexpected error: %v", err)
	}
	if len(list) != 1 || list[0].Username != "alice" {
		t.Fatalf("unexpected users result: %+v", list)
	}

	_, err = pg.GetUser(ctx, builder.NewUser().Usernames("missing"))
	if commonerrors.Errno(err) != consts.ErrnoUserNotFound {
		t.Fatalf("expected user not found errno, got err=%v", err)
	}

	err = pg.CreateUser(ctx, nil, &UserModel{Username: "alice", Password: "x"})
	if commonerrors.Errno(err) != consts.ErrnoUserAlreadyExists {
		t.Fatalf("expected user already exists errno, got err=%v", err)
	}
}

func TestPostgresContainerCRUD(t *testing.T) {
	pg := newTestPostgres(t)
	ctx := context.Background()

	create := &ContainerModel{
		UserID:    10,
		Name:      "demo",
		Image:     "busybox:latest",
		Command:   `["sleep","5"]`,
		Env:       `{"A":"1"}`,
		RuntimeID: "runtime-id",
		Status:    "created",
	}
	if err := pg.CreateContainer(ctx, nil, create); err != nil {
		t.Fatalf("CreateContainer unexpected error: %v", err)
	}
	if create.ID == 0 || create.CreatedAt.IsZero() || create.UpdatedAt.IsZero() {
		t.Fatalf("CreateContainer should hydrate id/timestamps, got %+v", create)
	}

	items, err := pg.BatchGetContainer(ctx, builder.NewContainer().UserIDs(10).Order("id DESC"))
	if err != nil {
		t.Fatalf("BatchGetContainer unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID == 0 {
		t.Fatalf("unexpected container list: %+v", items)
	}
	create.ID = items[0].ID

	got, err := pg.GetContainer(ctx, builder.NewContainer().IDs(create.ID).UserIDs(10))
	if err != nil {
		t.Fatalf("GetContainer unexpected error: %v", err)
	}
	if got.RuntimeID != "runtime-id" {
		t.Fatalf("unexpected runtime id: %+v", got)
	}

	create.Status = "running"
	create.Name = "demo-2"
	if err = pg.UpdateContainer(ctx, nil, create); err != nil {
		t.Fatalf("UpdateContainer unexpected error: %v", err)
	}

	got, err = pg.GetContainer(ctx, builder.NewContainer().IDs(create.ID).UserIDs(10))
	if err != nil {
		t.Fatalf("GetContainer after update unexpected error: %v", err)
	}
	if got.Status != "running" || got.Name != "demo-2" {
		t.Fatalf("unexpected updated container: %+v", got)
	}

	err = pg.UpdateContainer(ctx, nil, &ContainerModel{Model: gorm.Model{ID: 9999}, UserID: 10})
	if commonerrors.Errno(err) != consts.ErrnoContainerNotFound {
		t.Fatalf("expected container not found on update, got err=%v", err)
	}

	if err = pg.DeleteContainer(ctx, nil, create.ID, 10); err != nil {
		t.Fatalf("DeleteContainer unexpected error: %v", err)
	}

	err = pg.DeleteContainer(ctx, nil, create.ID, 10)
	if commonerrors.Errno(err) != consts.ErrnoContainerNotFound {
		t.Fatalf("expected container not found on delete, got err=%v", err)
	}

	_, err = pg.GetContainer(ctx, builder.NewContainer().IDs(create.ID).UserIDs(10))
	if commonerrors.Errno(err) != consts.ErrnoContainerNotFound {
		t.Fatalf("expected container not found on get, got err=%v", err)
	}
}

func TestPostgresCreateFile(t *testing.T) {
	pg := newTestPostgres(t)
	ctx := context.Background()

	create := &FileModel{
		UserID:        7,
		FileName:      "a.txt",
		ObjectKey:     "users/7/a.txt",
		ContentType:   "text/plain",
		Size:          3,
		WorkspacePath: "/tmp/a.txt",
	}
	if err := pg.CreateFile(ctx, nil, create); err != nil {
		t.Fatalf("CreateFile unexpected error: %v", err)
	}
	if create.ID == 0 || create.CreatedAt.IsZero() || create.UpdatedAt.IsZero() {
		t.Fatalf("CreateFile should hydrate id/timestamps, got %+v", create)
	}
	var count int64
	if err := pg.db.WithContext(ctx).Model(&FileModel{}).Where("user_id = ? AND file_name = ?", create.UserID, create.FileName).Count(&count).Error; err != nil {
		t.Fatalf("count created file failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one created file, got %d", count)
	}
}
