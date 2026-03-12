package file

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
}
