package command

import (
	"context"
	stderrors "errors"

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
	locker  ContainerActionLocker
}

func NewDeleteContainerHandler(
	repo container.Repository,
	runtime container.Runtime,
	logger *logrus.Logger,
	lockers ...ContainerActionLocker,
) DeleteContainerHandler {
	if repo == nil {
		panic("delete container's repository is nil")
	}
	if runtime == nil {
		panic("delete container's runtime is nil")
	}
	locker := defaultContainerActionLocker
	if len(lockers) > 0 {
		if lockers[0] == nil {
			panic("delete container's locker is nil")
		}
		locker = lockers[0]
	}

	return decorator.ApplyCommandDecorators(
		deleteContainerHandler{
			repo:    repo,
			runtime: runtime,
			locker:  locker,
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

	unlock, waited, err := h.locker.Lock(ctx, cmd.UserID, cmd.ContainerID)
	if err != nil {
		if stderrors.Is(err, ErrContainerActionWaitTimeout) {
			return nil, errors.NewWithMsgf(consts.ErrnoContainerActionWaitTimeout, "等待超時（資源已被佔用）請稍後重試")
		}
		return nil, err
	}
	defer unlock()

	data, err := h.repo.GetByIDAndUser(ctx, cmd.ContainerID, cmd.UserID)
	if err != nil {
		if waited && errors.Errno(err) == consts.ErrnoContainerNotFound {
			return nil, errors.NewWithMsgf(consts.ErrnoContainerNotFound, "container has been deleted by another request")
		}
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
