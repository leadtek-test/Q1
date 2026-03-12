package app

import "github.com/leadtek-test/q1/container/app/command"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateUser command.CreateUserHandler
	LoginUser  command.LoginUserHandler
}

type Queries struct{}
