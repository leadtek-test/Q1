package adapters

import (
	"fmt"
	"testing"
	"time"

	"github.com/leadtek-test/q1/container/infrastructure/persistent"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAdaptersTestPostgres(t *testing.T) *persistent.Postgres {
	t.Helper()

	dsn := fmt.Sprintf("file:container_adapters_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err = db.AutoMigrate(&persistent.UserModel{}, &persistent.FileModel{}, &persistent.ContainerModel{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	return persistent.NewPostgresWithDB(db)
}
