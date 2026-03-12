package adapters

import (
	"context"

	"github.com/leadtek-test/q1/container/domain/user"
	"github.com/leadtek-test/q1/container/infrastructure/persistent"
	"github.com/leadtek-test/q1/container/infrastructure/persistent/builder"
)

type UserRepositoryPostgres struct {
	db *persistent.Postgres
}

func NewUserRepositoryPostgres(db *persistent.Postgres) *UserRepositoryPostgres {
	return &UserRepositoryPostgres{
		db: db,
	}
}

func (u UserRepositoryPostgres) Create(ctx context.Context, usr *user.User) error {
	return u.db.CreateUser(ctx, nil, &persistent.UserModel{
		Username: usr.Username,
		Password: usr.PasswordHashed,
	})
}

func (u UserRepositoryPostgres) GetByUsername(ctx context.Context, username string) (user.User, error) {
	data, err := u.db.GetUser(ctx, builder.NewUser().Usernames(username))
	if err != nil {
		return user.User{}, err
	}
	return user.User{
		ID:             data.ID,
		Username:       data.Username,
		PasswordHashed: data.Password,
		CreatedAt:      data.CreatedAt,
		UpdatedAt:      data.UpdatedAt,
	}, nil
}

func (u UserRepositoryPostgres) GetByID(ctx context.Context, id uint) (user.User, error) {
	data, err := u.db.GetUser(ctx, builder.NewUser().IDs(id))
	if err != nil {
		return user.User{}, err
	}
	return user.User{
		ID:             data.ID,
		Username:       data.Username,
		PasswordHashed: data.Password,
		CreatedAt:      data.CreatedAt,
		UpdatedAt:      data.UpdatedAt,
	}, nil
}
