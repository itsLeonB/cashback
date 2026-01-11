//go:build job

package appembed

import "embed"

//go:embed internal/adapters/db/postgres/migrations/*.sql
var Migrations embed.FS

//go:embed assets/transfer-methods
var TransferMethodAssets embed.FS
