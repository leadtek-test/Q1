package adapters

import (
	"context"
	"testing"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/user"
)

func TestUserRepositoryPostgresCRUD(t *testing.T) {
	repo := NewUserRepositoryPostgres(newAdaptersTestPostgres(t))
	ctx := context.Background()

	target := &user.User{
		Username:       "alice",
		PasswordHashed: "hashed",
	}
	if err := repo.Create(ctx, target); err != nil {
		t.Fatalf("Create unexpected error: %v", err)
	}

	byUsername, err := repo.GetByUsername(ctx, "alice")
	if err != nil {
		t.Fatalf("GetByUsername unexpected error: %v", err)
	}
	if byUsername.Username != "alice" || byUsername.PasswordHashed != "hashed" || byUsername.ID == 0 {
		t.Fatalf("unexpected user by username: %+v", byUsername)
	}

	byID, err := repo.GetByID(ctx, byUsername.ID)
	if err != nil {
		t.Fatalf("GetByID unexpected error: %v", err)
	}
	if byID.ID != byUsername.ID || byID.Username != "alice" {
		t.Fatalf("unexpected user by id: %+v", byID)
	}
}

func TestUserRepositoryPostgresNotFound(t *testing.T) {
	repo := NewUserRepositoryPostgres(newAdaptersTestPostgres(t))

	_, err := repo.GetByUsername(context.Background(), "missing")
	if commonerrors.Errno(err) != consts.ErrnoUserNotFound {
		t.Fatalf("expected user not found errno, got err=%v", err)
	}
}
