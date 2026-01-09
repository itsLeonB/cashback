package debts

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type TransferMethod struct {
	crud.BaseEntity
	Name     string
	Display  string
	IconURL  sql.NullString
	ParentID uuid.NullUUID
}
