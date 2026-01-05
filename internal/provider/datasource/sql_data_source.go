package datasource

import (
	"database/sql"
	"fmt"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/ungerr"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ProvideAndConfigureSQL() (*gorm.DB, *sql.DB, error) {
	gormDB, err := gorm.Open(postgres.Open(dsn()), &gorm.Config{})
	if err != nil {
		return nil, nil, ungerr.Wrap(err, "error opening gorm connection")
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, nil, ungerr.Wrap(err, "error obtaining sql.DB instance")
	}

	sqlDB.SetMaxOpenConns(config.Global.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Global.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.Global.DB.ConnMaxLifetime)

	return gormDB, sqlDB, nil
}

func dsn() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		config.Global.DB.Host,
		config.Global.DB.User,
		config.Global.DB.Password,
		config.Global.DB.Name,
		config.Global.DB.Port,
	)
}
