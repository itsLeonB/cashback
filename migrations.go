package appembed

import "embed"

//go:embed internal/adapters/db/postgres/migrations/*.sql
var Migrations embed.FS
