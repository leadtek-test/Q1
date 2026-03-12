package service

import (
	"context"

	"github.com/leadtek-test/q1/container/adapters"
	"github.com/leadtek-test/q1/container/app"
	"github.com/leadtek-test/q1/container/app/command"
	"github.com/leadtek-test/q1/container/infrastructure/persistent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const defaultMaxUploadFileSize = int64(20 * 1024 * 1024)

// NewApplication 業務邏輯整合，回傳功能實體與關閉函式
func NewApplication(ctx context.Context) (app.Application, func()) {
	// TODO add GRPC client or MQ channel... etc
	return newApplication(ctx), nil
}

func newApplication(_ context.Context) app.Application {
	viper.SetDefault("file.max-size", defaultMaxUploadFileSize)
	viper.SetDefault("file.workspace-root", "./workspace")
	viper.SetDefault("file.object-root", "./object-storage")

	postgresDB := persistent.NewPostgres()
	userRepoPostgres := adapters.NewUserRepositoryPostgres(postgresDB)
	fileRepoPostgres := adapters.NewFileRepositoryPostgres(postgresDB)

	tokenManager := adapters.NewTokenManagerRepositoryJWT(
		viper.GetString("security.jwt-secret"),
		viper.GetDuration("security.jwt-expire-time"),
	)
	logger := logrus.StandardLogger()
	hasher := adapters.NewHasherRepositoryMD5()
	objectStorage := adapters.NewObjectStorageRepositoryLocal(viper.GetString("file.object-root"))
	workspace := adapters.NewWorkspaceRepositoryLocal(viper.GetString("file.workspace-root"))

	maxFileSize := viper.GetInt64("file.max-size")
	if maxFileSize <= 0 {
		maxFileSize = defaultMaxUploadFileSize
	}

	return app.Application{
		Commands: app.Commands{
			CreateUser: command.NewCreateUserHandler(userRepoPostgres, hasher, logger),
			LoginUser:  command.NewLoginUserHandler(userRepoPostgres, hasher, tokenManager, logger),
			UploadFile: command.NewUploadFileHandler(fileRepoPostgres, objectStorage, workspace, maxFileSize, logger),
		},
		Queries: app.Queries{},
	}
}
