package command

import (
	"context"
	"strings"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/decorator"
	"github.com/leadtek-test/q1/common/handler/errors"
	domainjob "github.com/leadtek-test/q1/container/domain/job"
	"github.com/sirupsen/logrus"
)

type CreateContainerJob struct {
	UserID  uint
	Name    string
	Image   string
	Command []string
	Env     map[string]string
}

type CreateContainerJobResult struct {
	JobID string
}

type CreateContainerJobHandler decorator.CommandHandler[CreateContainerJob, *CreateContainerJobResult]

type createContainerJobHandler struct {
	dispatcher domainjob.CreateContainerDispatcher
}

func NewCreateContainerJobHandler(
	dispatcher domainjob.CreateContainerDispatcher,
	logger *logrus.Logger,
) CreateContainerJobHandler {
	if dispatcher == nil {
		panic("create container job's dispatcher is nil")
	}

	return decorator.ApplyCommandDecorators(
		createContainerJobHandler{
			dispatcher: dispatcher,
		},
		logger,
	)
}

func (h createContainerJobHandler) Handle(ctx context.Context, cmd CreateContainerJob) (*CreateContainerJobResult, error) {
	task, err := h.validate(cmd)
	if err != nil {
		return nil, err
	}

	jobID, err := h.dispatcher.DispatchCreateContainer(ctx, task)
	if err != nil {
		return nil, err
	}

	return &CreateContainerJobResult{JobID: jobID}, nil
}

func (h createContainerJobHandler) validate(cmd CreateContainerJob) (domainjob.CreateContainerTask, error) {
	if cmd.UserID == 0 {
		return domainjob.CreateContainerTask{}, errors.New(consts.ErrnoAuthInvalidToken)
	}

	image := strings.TrimSpace(cmd.Image)
	if image == "" {
		return domainjob.CreateContainerTask{}, errors.New(consts.ErrnoContainerImageRequired)
	}

	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		name = "default"
	}

	commandData := cloneStrings(cmd.Command)
	if commandData == nil {
		commandData = make([]string, 0)
	}

	envData := cloneStringMap(cmd.Env)
	if envData == nil {
		envData = map[string]string{}
	}

	return domainjob.CreateContainerTask{
		UserID:  cmd.UserID,
		Name:    name,
		Image:   image,
		Command: commandData,
		Env:     envData,
	}, nil
}
