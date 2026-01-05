package datasource

import (
	"database/sql"
	"fmt"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/ungerr"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ProvideAndConfigureSQL(cfg config.DB) (*gorm.DB, *sql.DB, error) {
	gormDB, err := gorm.Open(postgres.Open(dsn(cfg)), &gorm.Config{})
	if err != nil {
		return nil, nil, ungerr.Wrap(err, "error opening gorm connection")
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, nil, ungerr.Wrap(err, "error obtaining sql.DB instance")
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err = sqlDB.Ping(); err != nil {
		return nil, nil, ungerr.Wrap(err, "error pinging SQL DB")
	}

	return gormDB, sqlDB, nil
}

func dsn(cfg config.DB) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
	)
}
