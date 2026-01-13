package debts

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type ParentFilter string

var (
	ParentOnly     ParentFilter = "parents"         // ParentID IS NULL
	ChildOnly      ParentFilter = "children"        // ParentID IS NOT NULL
	ForTransaction ParentFilter = "for-transaction" // Parents and Children that User Profile has (via ProfileTransferMethods)
	All            ParentFilter = "all"
)

type TransferMethod struct {
	crud.BaseEntity
	Name     string
	Display  string
	IconURL  sql.NullString
	ParentID uuid.NullUUID
}
