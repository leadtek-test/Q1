package adapters

import (
	"context"

	"github.com/leadtek-test/q1/container/domain/file"
	"github.com/leadtek-test/q1/container/infrastructure/persistent"
)

type FileRepositoryPostgres struct {
	db *persistent.Postgres
}

func NewFileRepositoryPostgres(db *persistent.Postgres) *FileRepositoryPostgres {
	return &FileRepositoryPostgres{db: db}
}

func (f FileRepositoryPostgres) Create(ctx context.Context, data *file.File) error {
	model := &persistent.FileModel{
		UserID:        data.UserID,
		FileName:      data.FileName,
		ObjectKey:     data.ObjectKey,
		ContentType:   data.ContentType,
		Size:          data.Size,
		WorkspacePath: data.WorkspacePath,
	}

	if err := f.db.CreateFile(ctx, nil, model); err != nil {
		return err
	}

	data.ID = model.ID
	data.CreatedAt = model.CreatedAt
	return nil
}
