package monetization

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type Subscription struct {
	crud.BaseEntity
	ProfileID     uuid.UUID
	PlanVersionID uuid.UUID
	EndsAt        sql.NullTime
	CanceledAt    sql.NullTime
	AutoRenew     bool

	// Relationships
	PlanVersion PlanVersion
}
