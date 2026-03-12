package adapters

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/handler/errors"
)

type ObjectStorageRepositoryLocal struct {
	root string
}

func NewObjectStorageRepositoryLocal(root string) *ObjectStorageRepositoryLocal {
	return &ObjectStorageRepositoryLocal{root: root}
}

func (o ObjectStorageRepositoryLocal) Upload(ctx context.Context, key string, body io.Reader, _ int64, _ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	relativePath := sanitizeRelativePath(key)
	if relativePath == "" {
		return errors.New(consts.ErrnoFileUploadFailed)
	}
	fullPath := filepath.Join(o.root, relativePath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return errors.NewWithError(consts.ErrnoFileUploadFailed, err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return errors.NewWithError(consts.ErrnoFileUploadFailed, err)
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err = io.Copy(file, body); err != nil {
		return errors.NewWithError(consts.ErrnoFileUploadFailed, err)
	}
	return nil
}

func sanitizeRelativePath(path string) string {
	clean := filepath.Clean("/" + strings.TrimSpace(path))
	return strings.TrimPrefix(clean, "/")
}
