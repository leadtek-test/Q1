package adapters

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/handler/errors"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domaincontainer "github.com/leadtek-test/q1/container/domain/container"
	domainfile "github.com/leadtek-test/q1/container/domain/file"
	domainjob "github.com/leadtek-test/q1/container/domain/job"
	"github.com/sirupsen/logrus"
)

const DefaultCreateContainerJobQueueSize = 128

type createContainerQueueTask struct {
	JobID string
	Task  domainjob.CreateContainerTask
}

type CreateContainerDispatcherChannel struct {
	jobRepo   domainjob.Repository
	repo      domaincontainer.Repository
	runtime   domaincontainer.Runtime
	workspace domainfile.Workspace
	logger    *logrus.Logger
	queue     chan createContainerQueueTask
}

var _ domainjob.CreateContainerDispatcher = (*CreateContainerDispatcherChannel)(nil)

func NewCreateContainerDispatcherChannel(
	jobRepo domainjob.Repository,
	repo domaincontainer.Repository,
	runtime domaincontainer.Runtime,
	workspace domainfile.Workspace,
	queueSize int,
	logger *logrus.Logger,
) *CreateContainerDispatcherChannel {
	if jobRepo == nil {
		panic("create container dispatcher's repository is nil")
	}
	if repo == nil {
		panic("create container dispatcher's container repository is nil")
	}
	if runtime == nil {
		panic("create container dispatcher's runtime is nil")
	}
	if workspace == nil {
		panic("create container dispatcher's workspace is nil")
	}
	if queueSize <= 0 {
		queueSize = DefaultCreateContainerJobQueueSize
	}
	return &CreateContainerDispatcherChannel{
		jobRepo:   jobRepo,
		repo:      repo,
		runtime:   runtime,
		workspace: workspace,
		logger:    logger,
		queue:     make(chan createContainerQueueTask, queueSize),
	}
}

func (d *CreateContainerDispatcherChannel) DispatchCreateContainer(ctx context.Context, task domainjob.CreateContainerTask) (string, error) {
	normalizedTask, err := d.normalizeTask(task)
	if err != nil {
		return "", err
	}

	jobID := uuid.NewString()
	data := &domainjob.CreateContainerJob{
		JobID:   jobID,
		UserID:  normalizedTask.UserID,
		Name:    normalizedTask.Name,
		Image:   normalizedTask.Image,
		Command: cloneStrings(normalizedTask.Command),
		Env:     cloneStringMap(normalizedTask.Env),
		Status:  domainjob.CreateContainerJobStatusAccepted,
	}
	if err = d.jobRepo.Create(ctx, data); err != nil {
		return "", err
	}

	queueTask := createContainerQueueTask{
		JobID: jobID,
		Task:  normalizedTask,
	}

	select {
	case d.queue <- queueTask:
		return jobID, nil
	default:
		err := commonerrors.New(consts.ErrnoContainerCreateJobQueueFull)
		d.markFailed(context.Background(), queueTask.JobID, queueTask.Task.UserID, err.Error())
		return "", err
	}
}

func (d *CreateContainerDispatcherChannel) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-d.queue:
			go d.process(task)
		}
	}
}

func (d *CreateContainerDispatcherChannel) process(task createContainerQueueTask) {
	processing := &domainjob.CreateContainerJob{
		JobID:        task.JobID,
		UserID:       task.Task.UserID,
		Status:       domainjob.CreateContainerJobStatusCreating,
		ErrorMessage: "",
		ContainerID:  0,
	}
	if err := d.jobRepo.Update(context.Background(), processing); err != nil {
		d.logError(err, "failed to update create-container job to creating", "job_id", task.JobID)
		return
	}

	containerID, err := d.executeCreateContainer(context.Background(), task.Task)
	if err != nil {
		d.markFailed(context.Background(), task.JobID, task.Task.UserID, err.Error())
		return
	}

	done := &domainjob.CreateContainerJob{
		JobID:        task.JobID,
		UserID:       task.Task.UserID,
		Status:       domainjob.CreateContainerJobStatusSucceeded,
		ErrorMessage: "",
		ContainerID:  containerID,
	}
	if err = d.jobRepo.Update(context.Background(), done); err != nil {
		d.logError(err, "failed to update create-container job to succeeded", "job_id", task.JobID)
	}
}

func (d *CreateContainerDispatcherChannel) markFailed(ctx context.Context, jobID string, userID uint, message string) {
	failed := &domainjob.CreateContainerJob{
		JobID:        jobID,
		UserID:       userID,
		Status:       domainjob.CreateContainerJobStatusFailed,
		ErrorMessage: message,
	}
	if err := d.jobRepo.Update(ctx, failed); err != nil {
		d.logError(err, "failed to update create-container job to failed", "job_id", jobID)
	}
}

func (d *CreateContainerDispatcherChannel) executeCreateContainer(ctx context.Context, task domainjob.CreateContainerTask) (uint, error) {
	workspacePath, err := d.workspace.EnsureUserDir(task.UserID)
	if err != nil {
		return 0, errors.NewWithError(consts.ErrnoContainerWorkspacePrepareFail, err)
	}

	runtimeID, err := d.runtime.Create(ctx, task.UserID, domaincontainer.CreateSpec{
		Name:    task.Name,
		Image:   task.Image,
		Command: cloneStrings(task.Command),
		Env:     cloneStringMap(task.Env),
	}, workspacePath)
	if err != nil {
		return 0, errors.NewWithError(consts.ErrnoContainerRuntimeCreateFail, err)
	}

	data := &domaincontainer.Container{
		UserID:    task.UserID,
		Name:      task.Name,
		Image:     task.Image,
		Command:   cloneStrings(task.Command),
		Env:       cloneStringMap(task.Env),
		RuntimeID: runtimeID,
		Status:    domaincontainer.StatusCreated,
	}
	if err = d.repo.Create(ctx, data); err != nil {
		return 0, err
	}

	return data.ID, nil
}

func (d *CreateContainerDispatcherChannel) normalizeTask(task domainjob.CreateContainerTask) (domainjob.CreateContainerTask, error) {
	if task.UserID == 0 {
		return domainjob.CreateContainerTask{}, errors.New(consts.ErrnoAuthInvalidToken)
	}

	image := strings.TrimSpace(task.Image)
	if image == "" {
		return domainjob.CreateContainerTask{}, errors.New(consts.ErrnoContainerImageRequired)
	}

	name := strings.TrimSpace(task.Name)
	if name == "" {
		name = "default"
	}

	commandData := cloneStrings(task.Command)
	if commandData == nil {
		commandData = make([]string, 0)
	}

	envData := cloneStringMap(task.Env)
	if envData == nil {
		envData = map[string]string{}
	}

	return domainjob.CreateContainerTask{
		UserID:  task.UserID,
		Name:    name,
		Image:   image,
		Command: commandData,
		Env:     envData,
	}, nil
}

func (d *CreateContainerDispatcherChannel) logError(err error, msg string, args ...any) {
	if d.logger == nil {
		return
	}
	entry := d.logger.WithError(err)
	if len(args)%2 == 0 {
		fields := logrus.Fields{}
		for i := 0; i < len(args); i += 2 {
			key, ok := args[i].(string)
			if !ok {
				continue
			}
			fields[key] = args[i+1]
		}
		entry = entry.WithFields(fields)
	}
	entry.Error(msg)
}

func cloneStrings(input []string) []string {
	if input == nil {
		return nil
	}
	output := make([]string, len(input))
	copy(output, input)
	return output
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}
