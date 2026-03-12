package container

import "context"

type Repository interface {
	Create(ctx context.Context, c *Container) error
	ListByUser(ctx context.Context, userID uint) ([]Container, error)
}
