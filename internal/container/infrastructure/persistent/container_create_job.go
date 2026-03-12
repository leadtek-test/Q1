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

type ContainerCreateJobModel struct {
	gorm.Model
	JobID        string `gorm:"size:64;uniqueIndex;not null"`
	UserID       uint   `gorm:"index;not null"`
	Name         string `gorm:"size:128;not null;default:''"`
	Image        string `gorm:"size:255;not null"`
	Command      string `gorm:"type:text;not null;default:'[]'"`
	Env          string `gorm:"type:text;not null;default:'{}'"`
	Status       string `gorm:"size:32;index;not null"`
	ErrorMessage string `gorm:"type:text;not null;default:''"`
	ContainerID  uint   `gorm:"not null;default:0"`
}

func (d Postgres) CreateContainerCreateJob(ctx context.Context, tx *gorm.DB, create *ContainerCreateJobModel) (err error) {
	_, deferLog := logging.WhenPostgres(ctx, "CreateContainerCreateJob", create)
	defer deferLog(create, &err)

	err = d.UseTransaction(tx).WithContext(ctx).Clauses(clause.Returning{}).Create(create).Error
	if err != nil {
		return errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return nil
}

func (d Postgres) GetContainerCreateJobByJobIDAndUser(ctx context.Context, jobID string, userID uint) (result *ContainerCreateJobModel, err error) {
	_, deferLog := logging.WhenPostgres(ctx, "GetContainerCreateJobByJobIDAndUser", jobID, userID)
	defer deferLog(result, &err)

	var data ContainerCreateJobModel
	err = d.db.WithContext(ctx).Where("job_id = ? AND user_id = ?", jobID, userID).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors2.New(consts.ErrnoContainerCreateJobNotFound)
		}
		return nil, errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return &data, nil
}

func (d Postgres) UpdateContainerCreateJob(ctx context.Context, tx *gorm.DB, update *ContainerCreateJobModel) (err error) {
	_, deferLog := logging.WhenPostgres(ctx, "UpdateContainerCreateJob", update)
	defer deferLog(update, &err)

	payload := map[string]any{
		"status":        update.Status,
		"error_message": update.ErrorMessage,
		"container_id":  update.ContainerID,
	}
	result := d.UseTransaction(tx).WithContext(ctx).Model(&ContainerCreateJobModel{}).
		Where("job_id = ? AND user_id = ?", update.JobID, update.UserID).
		Updates(payload)
	if result.Error != nil {
		return errors2.NewWithError(consts.ErrnoDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return errors2.New(consts.ErrnoContainerCreateJobNotFound)
	}
	return nil
}
