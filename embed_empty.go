//go:build !job

package appembed

import "embed"

var Migrations embed.FS
var TransferMethodAssets embed.FS
