package persistent

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	db *gorm.DB
}

func NewPostgres() *Postgres {
	viper.SetDefault("postgres.sslmode", "disable")
	viper.SetDefault("postgres.timezone", "Asia/Taipei")
	viper.SetDefault("postgres.max-open-conns", 25)
	viper.SetDefault("postgres.max-idle-conns", 10)
	viper.SetDefault("postgres.conn-max-lifetime", time.Hour)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		viper.GetString("postgres.host"),
		viper.GetString("postgres.user"),
		viper.GetString("postgres.password"),
		viper.GetString("postgres.dbname"),
		viper.GetInt("postgres.port"),
		viper.GetString("postgres.sslmode"),
		viper.GetString("postgres.timezone"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxOpenConns(viper.GetInt("postgres.max-open-conns"))
	sqlDB.SetMaxIdleConns(viper.GetInt("postgres.max-idle-conns"))
	sqlDB.SetConnMaxLifetime(viper.GetDuration("postgres.conn-max-lifetime"))

	if err = sqlDB.Ping(); err != nil {
		panic(err)
	}

	db.AutoMigrate(&UserModel{}, &FileModel{}, &ContainerModel{}, &ContainerCreateJobModel{})

	return &Postgres{db: db}
}

func NewPostgresWithDB(db *gorm.DB) *Postgres {
	return &Postgres{db: db}
}

func (d *Postgres) UseTransaction(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return d.db
	}
	return tx
}

func (d Postgres) StartTransaction(fc func(tx *gorm.DB) error) error {
	return d.db.Transaction(fc)
}
