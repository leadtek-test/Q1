package service

import (
	"context"

	"github.com/leadtek-test/q1/container/app"
)

// NewApplication 業務邏輯整合，回傳功能實體與關閉函式
func NewApplication(ctx context.Context) (app.Application, func()) {
	return newApplication(ctx), nil
}

func newApplication(_ context.Context) app.Application {
	return app.Application{
		Commands: app.Commands{},
		Queries:  app.Queries{},
	}
}
