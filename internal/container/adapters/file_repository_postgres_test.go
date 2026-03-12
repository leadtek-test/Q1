package adapters

import (
	"context"
	"testing"

	domainfile "github.com/leadtek-test/q1/container/domain/file"
	"github.com/leadtek-test/q1/container/infrastructure/persistent"
	"gorm.io/gorm"
)

func TestFileRepositoryPostgresCreate(t *testing.T) {
	pg := newAdaptersTestPostgres(t)
	repo := NewFileRepositoryPostgres(pg)
	ctx := context.Background()

	file := &domainfile.File{
		UserID:        2,
		FileName:      "a.txt",
		ObjectKey:     "users/2/a.txt",
		ContentType:   "text/plain",
		Size:          3,
		WorkspacePath: "/tmp/users/2/a.txt",
	}
	if err := repo.Create(ctx, file); err != nil {
		t.Fatalf("Create unexpected error: %v", err)
	}
	var count int64
	if err := pg.StartTransaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Model(&persistent.FileModel{}).Where("user_id = ? AND file_name = ?", file.UserID, file.FileName).Count(&count).Error
	}); err != nil {
		t.Fatalf("query created file failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one created file row, got %d", count)
	}
}
