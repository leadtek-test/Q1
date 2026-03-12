package service

import (
	"context"

	"github.com/leadtek-test/q1/container/adapters"
	"github.com/leadtek-test/q1/container/app"
	"github.com/leadtek-test/q1/container/app/command"
	"github.com/leadtek-test/q1/container/app/query"
	domainjob "github.com/leadtek-test/q1/container/domain/job"
	"github.com/leadtek-test/q1/container/infrastructure/persistent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const defaultMaxUploadFileSize = int64(20 * 1024 * 1024)

// NewApplication 業務邏輯整合，回傳功能實體、背景任務 listener 與關閉函式。
func NewApplication(ctx context.Context) (app.Application, domainjob.CreateContainerDispatcher, func()) {
	// TODO add GRPC client or MQ channel... etc
	return newApplication(ctx)
}

func newApplication(_ context.Context) (app.Application, domainjob.CreateContainerDispatcher, func()) {
	viper.SetDefault("file.max-size", defaultMaxUploadFileSize)
	viper.SetDefault("file.workspace-root", "./workspace")
	viper.SetDefault("file.workspace-runtime-root", "")
	viper.SetDefault("file.object-root", "./object-storage")
	viper.SetDefault("container.create-job-queue-size", adapters.DefaultCreateContainerJobQueueSize)

	postgresDB := persistent.NewPostgres()
	userRepoPostgres := adapters.NewUserRepositoryPostgres(postgresDB)
	fileRepoPostgres := adapters.NewFileRepositoryPostgres(postgresDB)
	containerRepoPostgres := adapters.NewContainerRepositoryPostgres(postgresDB)
	containerCreateJobRepoPostgres := adapters.NewContainerCreateJobRepositoryPostgres(postgresDB)

	tokenManager := adapters.NewTokenManagerRepositoryJWT(
		viper.GetString("security.jwt-secret"),
		viper.GetDuration("security.jwt-expire-time"),
	)
	logger := logrus.StandardLogger()
	hasher := adapters.NewHasherRepositoryMD5()
	objectStorage := adapters.NewObjectStorageRepositoryLocal(viper.GetString("file.object-root"))
	workspace := adapters.NewWorkspaceRepositoryLocal(viper.GetString("file.workspace-root"))
	containerRuntime, err := adapters.NewContainerRuntimeRepositoryDocker(
		viper.GetString("file.workspace-root"),
		viper.GetString("file.workspace-runtime-root"),
	)
	if err != nil {
		panic(err)
	}

	maxFileSize := viper.GetInt64("file.max-size")
	if maxFileSize <= 0 {
		maxFileSize = defaultMaxUploadFileSize
	}

	createContainerDispatcher := adapters.NewCreateContainerDispatcherChannel(
		containerCreateJobRepoPostgres,
		containerRepoPostgres,
		containerRuntime,
		workspace,
		viper.GetInt("container.create-job-queue-size"),
		logger,
	)
	containerActionLocker := command.NewInMemoryContainerActionLocker()

	return app.Application{
		Commands: app.Commands{
			CreateUser:            command.NewCreateUserHandler(userRepoPostgres, hasher, logger),
			LoginUser:             command.NewLoginUserHandler(userRepoPostgres, hasher, tokenManager, logger),
			UploadFile:            command.NewUploadFileHandler(fileRepoPostgres, objectStorage, workspace, maxFileSize, logger),
			CreateContainerJob:    command.NewCreateContainerJobHandler(createContainerDispatcher, logger),
			UpdateContainerStatus: command.NewUpdateContainerStatusHandler(containerRepoPostgres, containerRuntime, logger, containerActionLocker),
			DeleteContainer:       command.NewDeleteContainerHandler(containerRepoPostgres, containerRuntime, logger, containerActionLocker),
		},
		Queries: app.Queries{
			ListContainers:        query.NewListContainersHandler(containerRepoPostgres, containerRuntime, logger),
			GetCreateContainerJob: query.NewGetCreateContainerJobHandler(containerCreateJobRepoPostgres, logger),
		},
	}, createContainerDispatcher, nil
}
