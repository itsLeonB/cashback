package provider

import (
	"database/sql"
	"errors"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/provider/datasource"
	"github.com/itsLeonB/meq"
	"gorm.io/gorm"
)

type DataSources struct {
	Gorm  *gorm.DB
	SQL   *sql.DB
	Asynq meq.DB
}

func (ds *DataSources) Shutdown() error {
	var errs error
	if err := ds.SQL.Close(); err != nil {
		errs = errors.Join(errs, err)
	}
	if err := ds.Asynq.Shutdown(); err != nil {
		errs = errors.Join(errs, err)
	}
	return errs
}

func ProvideDataSource() (*DataSources, error) {
	gormDB, sqlDB, err := datasource.ProvideAndConfigureSQL(config.Global.DB)
	if err != nil {
		return nil, err
	}

	return &DataSources{
		Gorm:  gormDB,
		SQL:   sqlDB,
		Asynq: datasource.ProvideAsynq(config.Global.Valkey),
	}, nil
}
