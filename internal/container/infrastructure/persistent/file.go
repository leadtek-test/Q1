package persistent

import (
	"context"

	"github.com/leadtek-test/q1/common/consts"
	errors2 "github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/common/logging"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FileModel struct {
	gorm.Model
	UserID        uint   `gorm:"index;not null"`
	FileName      string `gorm:"size:255;not null"`
	ObjectKey     string `gorm:"size:512;index;not null"`
	ContentType   string `gorm:"size:128;not null"`
	Size          int64  `gorm:"not null"`
	WorkspacePath string `gorm:"size:512;not null"`
}

func (d Postgres) CreateFile(ctx context.Context, tx *gorm.DB, create *FileModel) (err error) {
	var returning FileModel
	_, deferLog := logging.WhenPostgres(ctx, "CreateFile", create)
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
