package query

import (
	"context"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/decorator"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/container"
	"github.com/sirupsen/logrus"
)

type ListContainers struct {
	UserID uint
}

type ContainerItem struct {
	ID        uint
	UserID    uint
	Name      string
	Image     string
	Command   []string
	Env       map[string]string
	RuntimeID string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListContainersResult struct {
	Containers []ContainerItem
}

type ListContainersHandler decorator.QueryHandler[ListContainers, *ListContainersResult]

type listContainersHandler struct {
	repo    container.Repository
	runtime container.Runtime
}

func NewListContainersHandler(repo container.Repository, runtime container.Runtime, logger *logrus.Logger) ListContainersHandler {
	if repo == nil {
		panic("list containers' repository is nil")
	}
	if runtime == nil {
		panic("list containers' runtime is nil")
	}

	return decorator.ApplyQueryDecorators(
		listContainersHandler{
			repo:    repo,
			runtime: runtime,
		},
		logger,
	)
}

func (h listContainersHandler) Handle(ctx context.Context, query ListContainers) (*ListContainersResult, error) {
	if query.UserID == 0 {
		return nil, errors.New(consts.ErrnoAuthInvalidToken)
	}

	containers, err := h.repo.ListByUser(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	items := make([]ContainerItem, 0, len(containers))
	for _, item := range containers {
		items = append(items, ContainerItem{
			ID:        item.ID,
			UserID:    item.UserID,
			Name:      item.Name,
			Image:     item.Image,
			Command:   item.Command,
			Env:       item.Env,
			RuntimeID: item.RuntimeID,
			Status:    string(item.Status),
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	return &ListContainersResult{
		Containers: items,
	}, nil
}
