package command

import (
	"context"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/decorator"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/container"
	"github.com/sirupsen/logrus"
)

type DeleteContainer struct {
	UserID      uint
	ContainerID uint
}

type DeleteContainerResult struct {
	Deleted bool `json:"deleted"`
}

type DeleteContainerHandler decorator.CommandHandler[DeleteContainer, *DeleteContainerResult]

type deleteContainerHandler struct {
	repo    container.Repository
	runtime container.Runtime
}

func NewDeleteContainerHandler(
	repo container.Repository,
	runtime container.Runtime,
	logger *logrus.Logger,
) DeleteContainerHandler {
	if repo == nil {
		panic("delete container's repository is nil")
	}
	if runtime == nil {
		panic("delete container's runtime is nil")
	}

	return decorator.ApplyCommandDecorators(
		deleteContainerHandler{
			repo:    repo,
			runtime: runtime,
		},
		logger,
	)
}

func (h deleteContainerHandler) Handle(ctx context.Context, cmd DeleteContainer) (*DeleteContainerResult, error) {
	if cmd.UserID == 0 {
		return nil, errors.New(consts.ErrnoAuthInvalidToken)
	}
	if cmd.ContainerID == 0 {
		return nil, errors.NewWithMsgf(consts.ErrnoRequestValidateError, "invalid container id")
	}

	data, err := h.repo.GetByIDAndUser(ctx, cmd.ContainerID, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if err = h.runtime.Delete(ctx, data.RuntimeID); err != nil {
		return nil, errors.NewWithError(consts.ErrnoContainerRuntimeDeleteFail, err)
	}

	if err = h.repo.Delete(ctx, cmd.ContainerID, cmd.UserID); err != nil {
		return nil, err
	}

	return &DeleteContainerResult{Deleted: true}, nil
}
