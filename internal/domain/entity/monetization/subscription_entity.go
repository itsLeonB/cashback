package monetization

import (
	"database/sql"
	"time"

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

func (s *Subscription) IsActive(t time.Time) bool {
	return !(s.EndsAt.Valid && s.EndsAt.Time.Before(t)) &&
		!(s.CanceledAt.Valid && s.CanceledAt.Time.Before(t))
}
