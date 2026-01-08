package debts

import (
	"github.com/itsLeonB/go-crud"
)

type TransferMethod struct {
	crud.BaseEntity
	Name    string
	Display string
}
