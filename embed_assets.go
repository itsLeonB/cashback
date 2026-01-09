//go:build asseter

package appembed

import "embed"

//go:embed assets/transfer-methods
var TransferMethodAssets embed.FS
