package persistent

import (
	"context"
	"errors"

	"github.com/leadtek-test/q1/common/consts"
	errors2 "github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/common/logging"
	"github.com/leadtek-test/q1/container/infrastructure/persistent/builder"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContainerModel struct {
	gorm.Model
	UserID    uint   `gorm:"index;not null"`
	Name      string `gorm:"size:128;not null"`
	Image     string `gorm:"size:255;not null"`
	Command   string `gorm:"type:text;not null;default:'[]'"`
	Env       string `gorm:"type:text;not null;default:'{}'"`
	RuntimeID string `gorm:"size:128;not null"`
	Status    string `gorm:"size:32;not null"`
}

func (d Postgres) CreateContainer(ctx context.Context, tx *gorm.DB, create *ContainerModel) (err error) {
	_, deferLog := logging.WhenPostgres(ctx, "CreateContainer", create)
	defer deferLog(create, &err)

	err = d.UseTransaction(tx).WithContext(ctx).Clauses(clause.Returning{}).Create(create).Error
	if err != nil {
		return errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return nil
}

func (d Postgres) GetContainer(ctx context.Context, query *builder.Container) (result *ContainerModel, err error) {
	_, deferLog := logging.WhenPostgres(ctx, "GetContainer", query)
	defer deferLog(result, &err)

	err = query.Fill(d.db.WithContext(ctx)).First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors2.New(consts.ErrnoContainerNotFound)
		}
		return nil, errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return result, nil
}

func (d Postgres) BatchGetContainer(ctx context.Context, query *builder.Container) (results []ContainerModel, err error) {
	_, deferLog := logging.WhenPostgres(ctx, "BatchGetContainer", query)
	defer deferLog(results, &err)

	err = query.Fill(d.db.WithContext(ctx)).Find(&results).Error
	if err != nil {
		return nil, errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return results, nil
}

func (d Postgres) UpdateContainer(ctx context.Context, tx *gorm.DB, update *ContainerModel) (err error) {
	where := builder.NewContainer().IDs(update.ID).UserIDs(update.UserID)
	_, deferLog := logging.WhenPostgres(ctx, "UpdateContainer", where, update)
	defer deferLog(update, &err)

	payload := map[string]any{
		"name":       update.Name,
		"image":      update.Image,
		"command":    update.Command,
		"env":        update.Env,
		"runtime_id": update.RuntimeID,
		"status":     update.Status,
	}
	result := where.Fill(d.UseTransaction(tx).WithContext(ctx).Model(&ContainerModel{})).Updates(payload)
	if result.Error != nil {
		return errors2.NewWithError(consts.ErrnoDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return errors2.New(consts.ErrnoContainerNotFound)
	}
	return nil
}

func (d Postgres) DeleteContainer(ctx context.Context, tx *gorm.DB, id, userID uint) (err error) {
	where := builder.NewContainer().IDs(id).UserIDs(userID)
	_, deferLog := logging.WhenPostgres(ctx, "DeleteContainer", where)
	defer deferLog(nil, &err)

	result := where.Fill(d.UseTransaction(tx).WithContext(ctx)).Delete(&ContainerModel{})
	if result.Error != nil {
		return errors2.NewWithError(consts.ErrnoDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return errors2.New(consts.ErrnoContainerNotFound)
	}
	return nil
}
