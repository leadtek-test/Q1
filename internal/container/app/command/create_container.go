package command

import (
	"context"
	"strings"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/decorator"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/container"
	"github.com/leadtek-test/q1/container/domain/file"
	"github.com/sirupsen/logrus"
)

type CreateContainer struct {
	UserID  uint
	Name    string
	Image   string
	Command []string
	Env     map[string]string
}

type CreateContainerResult struct {
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

type CreateContainerHandler decorator.CommandHandler[CreateContainer, *CreateContainerResult]

type createContainerHandler struct {
	repo      container.Repository
	runtime   container.Runtime
	workspace file.Workspace
}

func NewCreateContainerHandler(
	repo container.Repository,
	runtime container.Runtime,
	workspace file.Workspace,
	logger *logrus.Logger,
) CreateContainerHandler {
	if repo == nil {
		panic("create container's repository is nil")
	}
	if runtime == nil {
		panic("create container's runtime is nil")
	}
	if workspace == nil {
		panic("create container's workspace is nil")
	}

	return decorator.ApplyCommandDecorators(
		createContainerHandler{
			repo:      repo,
			runtime:   runtime,
			workspace: workspace,
		},
		logger,
	)
}

func (h createContainerHandler) Handle(ctx context.Context, cmd CreateContainer) (*CreateContainerResult, error) {
	spec, err := h.validate(cmd)
	if err != nil {
		return nil, err
	}

	workspacePath, err := h.workspace.EnsureUserDir(cmd.UserID)
	if err != nil {
		return nil, errors.NewWithError(consts.ErrnoContainerWorkspacePrepareFail, err)
	}

	runtimeID, err := h.runtime.Create(ctx, cmd.UserID, spec, workspacePath)
	if err != nil {
		return nil, errors.NewWithError(consts.ErrnoContainerRuntimeCreateFail, err)
	}

	data := &container.Container{
		UserID:    cmd.UserID,
		Name:      spec.Name,
		Image:     spec.Image,
		Command:   spec.Command,
		Env:       spec.Env,
		RuntimeID: runtimeID,
		Status:    container.StatusCreated,
	}

	if err = h.repo.Create(ctx, data); err != nil {
		return nil, err
	}

	return &CreateContainerResult{
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
	}, nil
}

func (h createContainerHandler) validate(cmd CreateContainer) (container.CreateSpec, error) {
	if cmd.UserID == 0 {
		return container.CreateSpec{}, errors.New(consts.ErrnoAuthInvalidToken)
	}

	image := strings.TrimSpace(cmd.Image)
	if image == "" {
		return container.CreateSpec{}, errors.New(consts.ErrnoContainerImageRequired)
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		name = "default"
	}

	commandData := cmd.Command
	if commandData == nil {
		commandData = make([]string, 0)
	}

	envData := cmd.Env
	if envData == nil {
		envData = map[string]string{}
	}

	return container.CreateSpec{
		Name:    name,
		Image:   image,
		Command: commandData,
		Env:     envData,
	}, nil
}
