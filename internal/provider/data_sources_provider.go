package provider

import (
	"database/sql"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/provider/datasource"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type DataSources struct {
	Gorm *gorm.DB
	SQL  *sql.DB
}

func (ds *DataSources) Shutdown() error {
	if err := ds.SQL.Close(); err != nil {
		return ungerr.Wrap(err, "error closing SQL db")
	}
	return nil
}

func ProvideDataSource() (*DataSources, error) {
	gormDB, sqlDB, err := datasource.ProvideAndConfigureSQL(config.Global.DB)
	if err != nil {
		return nil, err
	}

	return &DataSources{
		Gorm: gormDB,
		SQL:  sqlDB,
	}, nil
}
