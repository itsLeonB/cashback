package expenses

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type FeeCalculationMethod string

const (
	EqualSplitFee    FeeCalculationMethod = "EQUAL_SPLIT"
	ItemizedSplitFee FeeCalculationMethod = "ITEMIZED_SPLIT"
)

type OtherFee struct {
	crud.BaseEntity
	GroupExpenseID    uuid.UUID
	Name              string
	Amount            decimal.Decimal
	CalculationMethod FeeCalculationMethod
	Participants      []FeeParticipant `gorm:"foreignKey:OtherFeeID"`
}

func (of OtherFee) ProfileIDs() []uuid.UUID {
	return ezutil.MapSlice(of.Participants, func(fp FeeParticipant) uuid.UUID { return fp.ProfileID })
}

func (of OtherFee) TableName() string {
	return "group_expense_other_fees"
}
