package persistent

import (
	"context"
	"errors"

	"github.com/leadtek-test/q1/common/consts"
	errors2 "github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/common/logging"
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
	var returning ContainerModel
	_, deferLog := logging.WhenPostgres(ctx, "CreateContainer", create)
	defer deferLog(returning, &err)

	err = d.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(create).Error
	if err != nil {
		return errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}

	create.ID = returning.ID
	create.CreatedAt = returning.CreatedAt
	create.UpdatedAt = returning.UpdatedAt
	return nil
}

func (d Postgres) BatchGetContainerByUser(ctx context.Context, userID uint) (results []ContainerModel, err error) {
	_, deferLog := logging.WhenPostgres(ctx, "BatchGetContainerByUser", userID)
	defer deferLog(results, &err)

	err = d.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&results).Error
	if err != nil {
		return nil, errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return results, nil
}

func (d Postgres) GetContainerByIDAndUser(ctx context.Context, id, userID uint) (result *ContainerModel, err error) {
	_, deferLog := logging.WhenPostgres(ctx, "GetContainerByIDAndUser", id, userID)
	defer deferLog(result, &err)

	var data ContainerModel
	err = d.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors2.New(consts.ErrnoContainerNotFound)
		}
		return nil, errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return &data, nil
}

func (d Postgres) UpdateContainer(ctx context.Context, tx *gorm.DB, update *ContainerModel) (err error) {
	_, deferLog := logging.WhenPostgres(ctx, "UpdateContainer", update)
	defer deferLog(update, &err)

	payload := map[string]any{
		"name":       update.Name,
		"image":      update.Image,
		"command":    update.Command,
		"env":        update.Env,
		"runtime_id": update.RuntimeID,
		"status":     update.Status,
	}
	result := d.UseTransaction(tx).WithContext(ctx).Model(&ContainerModel{}).
		Where("id = ? AND user_id = ?", update.ID, update.UserID).
		Updates(payload)
	if result.Error != nil {
		return errors2.NewWithError(consts.ErrnoDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return errors2.New(consts.ErrnoContainerNotFound)
	}
	return nil
}

func (d Postgres) DeleteContainer(ctx context.Context, tx *gorm.DB, id, userID uint) (err error) {
	_, deferLog := logging.WhenPostgres(ctx, "DeleteContainer", id, userID)
	defer deferLog(nil, &err)

	result := d.UseTransaction(tx).WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&ContainerModel{})
	if result.Error != nil {
		return errors2.NewWithError(consts.ErrnoDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return errors2.New(consts.ErrnoContainerNotFound)
	}
	return nil
}
