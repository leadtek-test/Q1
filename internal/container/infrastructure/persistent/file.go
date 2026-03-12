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
	_, deferLog := logging.WhenPostgres(ctx, "CreateFile", create)
	defer deferLog(create, &err)

	err = d.UseTransaction(tx).WithContext(ctx).Clauses(clause.Returning{}).Create(create).Error
	if err != nil {
		return errors2.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return nil
}
