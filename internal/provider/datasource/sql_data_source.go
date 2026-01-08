package datasource

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/ungerr"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	sqlInstance *sqlConnection
	sqlOnce     sync.Once
)

type sqlConnection struct {
	gormDB *gorm.DB
	sqlDB  *sql.DB
}

func ProvideAndConfigureSQL(cfg config.DB) (*gorm.DB, *sql.DB, error) {
	var err error
	sqlOnce.Do(func() {
		gormDB, e := gorm.Open(postgres.Open(dsn(cfg)), &gorm.Config{})
		if e != nil {
			err = ungerr.Wrap(e, "error opening gorm connection")
			return
		}

		sqlDB, e := gormDB.DB()
		if e != nil {
			err = ungerr.Wrap(e, "error obtaining sql.DB instance")
			return
		}

		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		if e = sqlDB.Ping(); e != nil {
			err = ungerr.Wrap(e, "error pinging SQL DB")
			return
		}

		sqlInstance = &sqlConnection{gormDB: gormDB, sqlDB: sqlDB}
	})

	if err != nil {
		return nil, nil, err
	}

	return sqlInstance.gormDB, sqlInstance.sqlDB, nil
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
