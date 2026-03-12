package command

import (
	"context"
	"strings"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/decorator"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/auth"
	"github.com/leadtek-test/q1/container/domain/user"
	"github.com/sirupsen/logrus"
)

type CreateUser struct {
	Username string
	Password string
}

type CreateUserResult struct {
	UserID   uint
	Username string
}

type CreateUserHandler decorator.CommandHandler[CreateUser, *CreateUserResult]

type createUserHandler struct {
	userRepo user.Repository
	hasher   auth.PasswordHasher
}

func NewCreateUserHandler(userRepo user.Repository, hasher auth.PasswordHasher, logger *logrus.Logger) CreateUserHandler {
	if userRepo == nil {
		panic("create user's user repository is nil")
	}
	if hasher == nil {
		panic("create user's user hasher is nil")
	}

	return decorator.ApplyCommandDecorators(
		createUserHandler{
			userRepo: userRepo,
			hasher:   hasher,
		},
		logger,
	)
}

func (c createUserHandler) Handle(ctx context.Context, cmd CreateUser) (*CreateUserResult, error) {
	username := strings.TrimSpace(cmd.Username)
	password := strings.TrimSpace(cmd.Password)

	err := c.validate(ctx, username, password)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Username:       username,
		PasswordHashed: c.hasher.Hash(password),
	}

	if err = c.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return &CreateUserResult{UserID: u.ID, Username: u.Username}, nil
}

func (c createUserHandler) validate(ctx context.Context, username string, password string) error {
	if username == "" {
		return errors.New(consts.ErrnoUserUsernameRequired)
	}

	if password == "" {
		return errors.New(consts.ErrnoUserPasswordRequired)
	}

	if len(password) < 6 {
		return errors.New(consts.ErrnoUserPasswordTooShort)
	}

	_, err := c.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return errors.New(consts.ErrnoUserAlreadyExists)
	}

	if errors.Errno(err) != consts.ErrnoUserNotFound {
		return err
	}

	return nil
}
