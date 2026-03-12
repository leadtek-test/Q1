package command

import (
	"context"
	stderrors "errors"
	"strings"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/decorator"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/container"
	"github.com/sirupsen/logrus"
)

type UpdateContainerStatus struct {
	UserID      uint
	ContainerID uint
	Action      string
}

type UpdateContainerStatusResult struct {
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

type UpdateContainerStatusHandler decorator.CommandHandler[UpdateContainerStatus, *UpdateContainerStatusResult]

type updateContainerStatusHandler struct {
	repo    container.Repository
	runtime container.Runtime
	locker  ContainerActionLocker
}

func NewUpdateContainerStatusHandler(
	repo container.Repository,
	runtime container.Runtime,
	logger *logrus.Logger,
	lockers ...ContainerActionLocker,
) UpdateContainerStatusHandler {
	if repo == nil {
		panic("update container status's repository is nil")
	}
	if runtime == nil {
		panic("update container status's runtime is nil")
	}
	locker := defaultContainerActionLocker
	if len(lockers) > 0 {
		if lockers[0] == nil {
			panic("update container status's locker is nil")
		}
		locker = lockers[0]
	}

	return decorator.ApplyCommandDecorators(
		updateContainerStatusHandler{
			repo:    repo,
			runtime: runtime,
			locker:  locker,
		},
		logger,
	)
}

func (h updateContainerStatusHandler) Handle(ctx context.Context, cmd UpdateContainerStatus) (*UpdateContainerStatusResult, error) {
	action, err := normalizeContainerStatusAction(cmd.Action)
	if err != nil {
		return nil, err
	}
	if cmd.UserID == 0 {
		return nil, errors.New(consts.ErrnoAuthInvalidToken)
	}
	if cmd.ContainerID == 0 {
		return nil, errors.NewWithMsgf(consts.ErrnoRequestValidateError, "invalid container id")
	}

	unlock, waited, err := h.locker.Lock(ctx, cmd.UserID, cmd.ContainerID)
	if err != nil {
		if stderrors.Is(err, ErrContainerActionWaitTimeout) {
			return nil, errors.NewWithMsgf(consts.ErrnoContainerActionWaitTimeout, "waiting timeout(resource has been used)")
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

	switch action {
	case "start":
		// idempotent: already running.
		if data.Status == container.StatusRunning {
			return toUpdateContainerStatusResult(data), nil
		}
		if err = h.runtime.Start(ctx, data.RuntimeID); err != nil {
			return nil, errors.NewWithError(consts.ErrnoContainerRuntimeStartFail, err)
		}
		data.Status = container.StatusRunning
		if err = h.repo.Update(ctx, &data); err != nil {
			return nil, err
		}
	case "stop":
		// idempotent: anything not running is already "closed".
		if data.Status != container.StatusRunning {
			return toUpdateContainerStatusResult(data), nil
		}
		if err = h.runtime.Stop(ctx, data.RuntimeID); err != nil {
			return nil, errors.NewWithError(consts.ErrnoContainerRuntimeStopFail, err)
		}
		data.Status = container.StatusStopped
		if err = h.repo.Update(ctx, &data); err != nil {
			return nil, err
		}
	}

	return toUpdateContainerStatusResult(data), nil
}

func normalizeContainerStatusAction(action string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "start", "running":
		return "start", nil
	case "stop", "stopped":
		return "stop", nil
	default:
		return "", errors.New(consts.ErrnoContainerInvalidStatusAction)
	}
}

func toUpdateContainerStatusResult(data container.Container) *UpdateContainerStatusResult {
	return &UpdateContainerStatusResult{
		ID:        data.ID,
		UserID:    data.UserID,
		Name:      data.Name,
		Image:     data.Image,
		Command:   data.Command,
		Env:       data.Env,
		RuntimeID: data.RuntimeID,
		Status:    string(data.Status),
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}
}
