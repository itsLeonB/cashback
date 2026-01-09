package migrator

import (
	appembed "github.com/itsLeonB/cashback"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/provider/datasource"
	"github.com/itsLeonB/ungerr"
	"github.com/pressly/goose/v3"
)

func Run() error {
	_, db, err := datasource.ProvideAndConfigureSQL(config.Global.DB)
	if err != nil {
		return err
	}

	goose.SetBaseFS(appembed.Migrations)
	goose.SetLogger(logger.Global)

	if err = goose.SetDialect("postgres"); err != nil {
		return ungerr.Wrap(err, "error setting migrator dialect to postgres")
	}

	if err = goose.Up(db, "internal/adapters/db/postgres/migrations"); err != nil {
		return ungerr.Wrap(err, "error running migrations")
	}

	return nil
}
