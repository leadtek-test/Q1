package persistent

import (
	"context"
	"errors"
	"strings"

	"github.com/leadtek-test/q1/common/consts"
	errors2 "github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/common/logging"
	"github.com/leadtek-test/q1/container/infrastructure/persistent/builder"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserModel struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}

func (d Postgres) GetUser(ctx context.Context, query *builder.User) (result *UserModel, err error) {
	_, deferLog := logging.WhenPostgres(ctx, "GetUser", query)
	defer deferLog(result, &err)

	err = query.Fill(d.db.WithContext(ctx)).First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors2.New(consts.ErrnoUserNotFound)
		}
		return nil, errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return result, nil
}

func (d Postgres) BatchGetUser(ctx context.Context, query *builder.User) (results []UserModel, err error) {
	_, deferLog := logging.WhenPostgres(ctx, "BatchGetUser", query)
	defer deferLog(results, &err)

	err = query.Fill(d.db.WithContext(ctx)).Find(&results).Error
	if err != nil {
		return nil, errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return results, nil
}

func (d Postgres) CreateUser(ctx context.Context, tx *gorm.DB, create *UserModel) (err error) {
	_, deferLog := logging.WhenPostgres(ctx, "CreateUser", create)
	defer deferLog(create, &err)

	err = d.UseTransaction(tx).WithContext(ctx).Clauses(clause.Returning{}).Create(create).Error
	if err != nil {
		if isUniqueViolation(err) {
			return errors2.New(consts.ErrnoUserAlreadyExists)
		}
		return errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "UNIQUE constraint failed"))
}
