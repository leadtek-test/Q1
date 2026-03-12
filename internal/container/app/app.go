package app

import (
	"github.com/leadtek-test/q1/container/app/command"
	"github.com/leadtek-test/q1/container/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateUser            command.CreateUserHandler
	LoginUser             command.LoginUserHandler
	UploadFile            command.UploadFileHandler
	CreateContainerJob    command.CreateContainerJobHandler
	UpdateContainerStatus command.UpdateContainerStatusHandler
	DeleteContainer       command.DeleteContainerHandler
}

type Queries struct {
	ListContainers        query.ListContainersHandler
	GetCreateContainerJob query.GetCreateContainerJobHandler
}
