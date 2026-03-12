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

// NewApplication 業務邏輯整合，回傳功能實體與關閉函式
func NewApplication(ctx context.Context) (app.Application, func()) {
	// TODO add GRPC client or MQ channel... etc
	return newApplication(ctx), nil
}

func newApplication(_ context.Context) app.Application {
	postgresDB := persistent.NewPostgres()
	userRepoPostgres := adapters.NewUserRepositoryPostgres(postgresDB)

	tokenManager := adapters.NewTokenManagerRepositoryJWT(
		viper.GetString("security.jwt-secret"),
		viper.GetDuration("security.jwt-expire-time"),
	)
	hasher := adapters.NewHasherRepositoryMD5()
	return app.Application{
		Commands: app.Commands{
			CreateUser: command.NewCreateUserHandler(userRepoPostgres, hasher, tokenManager, logrus.StandardLogger()),
		},
		Queries: app.Queries{},
	}
}
