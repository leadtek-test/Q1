package command

import (
	"context"
	"strings"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/decorator"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/auth"
	"github.com/leadtek-test/q1/container/domain/user"
	"github.com/sirupsen/logrus"
)

type LoginUser struct {
	Username string
	Password string
}

type LoginUserResult struct {
	UserID    uint
	Username  string
	Token     string
	ExpiresAt time.Time
}

type LoginUserHandler decorator.CommandHandler[LoginUser, *LoginUserResult]

type loginUserHandler struct {
	userRepo user.Repository
	hasher   auth.PasswordHasher
	token    auth.TokenManager
}

func NewLoginUserHandler(userRepo user.Repository, hasher auth.PasswordHasher, token auth.TokenManager, logger *logrus.Logger) LoginUserHandler {
	if userRepo == nil {
		panic("create user's user repository is nil")
	}
	if hasher == nil {
		panic("create user's user hasher is nil")
	}
	if token == nil {
		panic("create user's user token is nil")
	}

	return decorator.ApplyCommandDecorators(
		loginUserHandler{
			userRepo: userRepo,
			hasher:   hasher,
			token:    token,
		},
		logger,
	)
}

func (l loginUserHandler) Handle(ctx context.Context, cmd LoginUser) (*LoginUserResult, error) {
	u, err := l.validate(ctx, cmd)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := l.token.Generate(u.UserID, u.Username)
	if err != nil {
		return nil, err
	}
	u.Token = token
	u.ExpiresAt = expiresAt

	return u, nil
}

func (l loginUserHandler) validate(ctx context.Context, cmd LoginUser) (*LoginUserResult, error) {
	username := strings.TrimSpace(cmd.Username)
	password := strings.TrimSpace(cmd.Password)
	if username == "" {
		return nil, errors.New(consts.ErrnoUserUsernameRequired)
	}

	if password == "" {
		return nil, errors.New(consts.ErrnoUserPasswordRequired)
	}

	if len(password) < 6 {
		return nil, errors.New(consts.ErrnoUserPasswordTooShort)
	}

	u, err := l.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return nil, errors.New(consts.ErrnoUserAlreadyExists)
	}

	if errors.Errno(err) != consts.ErrnoUserNotFound {
		return nil, err
	}

	if ok := l.hasher.Compare(password, u.PasswordHashed); !ok {
		return nil, errors.New(consts.ErrnoUserPasswordNotMatch)
	}

	return &LoginUserResult{UserID: u.ID, Username: u.Username}, nil
}
