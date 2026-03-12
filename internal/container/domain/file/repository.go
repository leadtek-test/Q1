package file

import "context"

type Repository interface {
	Create(ctx context.Context, f *File) error
}
