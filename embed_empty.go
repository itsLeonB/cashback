//go:build !migrator && !asseter

package appembed

import "embed"

var Migrations embed.FS
var TransferMethodAssets embed.FS
