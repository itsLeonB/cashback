package monetization

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
	"golang.org/x/text/currency"
)

type Plan struct {
	crud.BaseEntity
	Name     string
	IsActive bool
}

type BillingInterval string

const (
	MonthlyInterval BillingInterval = "monthly"
	YearlyInterval  BillingInterval = "yearly"
)

type PlanVersion struct {
	crud.BaseEntity
	PlanID             uuid.UUID
	PriceAmount        decimal.Decimal
	PriceCurrency      currency.Unit
	BillingInterval    BillingInterval
	BillUploadsDaily   uint
	BillUploadsMonthly uint
	EffectiveFrom      time.Time
	EffectiveTo        sql.NullTime
	IsDefault          bool

	// Relationships
	Plan Plan
}
