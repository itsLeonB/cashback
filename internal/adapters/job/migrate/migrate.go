package migrate

import (
	"database/sql"

	appembed "github.com/itsLeonB/cashback"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
	"github.com/pressly/goose/v3"
)

type Migrate struct {
	db *sql.DB
}

func Setup(providers *provider.Providers) (*Migrate, error) {
	goose.SetBaseFS(appembed.Migrations)
	goose.SetLogger(ezutil.NewSimpleLogger(config.Global.ServiceName, true, 0))

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, ungerr.Wrap(err, "error setting migrator dialect to postgres")
	}

	return &Migrate{providers.SQL}, nil
}

func (m *Migrate) Run() error {
	if err := goose.Up(m.db, "internal/adapters/db/postgres/migrations"); err != nil {
		return ungerr.Wrap(err, "error running migrations")
	}
	return nil
}
