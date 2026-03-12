package container

import "context"

type Repository interface {
	Create(ctx context.Context, c *Container) error
	GetByIDAndUser(ctx context.Context, id, userID uint) (Container, error)
	Update(ctx context.Context, c *Container) error
	Delete(ctx context.Context, id, userID uint) error
	ListByUser(ctx context.Context, userID uint) ([]Container, error)
}
