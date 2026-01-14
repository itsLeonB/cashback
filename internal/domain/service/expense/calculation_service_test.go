package expense_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/service/expense"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculationServiceRecalculateExpense(t *testing.T) {
	svc := expense.NewCalculationService()

	type args struct {
		expense       expenses.GroupExpense
		amountChanged bool
	}

	tests := []struct {
		name            string
		args            args
		wantItemsTotal  decimal.Decimal
		wantTotalAmount decimal.Decimal
		wantStatus      expenses.ExpenseStatus
		wantChanged     bool
		wantErr         bool
	}{
		{
			name: "amountChanged=false, no items, stays draft",
			args: args{
				expense: expenses.GroupExpense{
					Status: expenses.DraftExpense,
					Items:  []expenses.ExpenseItem{},
				},
				amountChanged: false,
			},
			wantItemsTotal:  decimal.Zero,
			wantTotalAmount: decimal.Zero,
			wantStatus:      expenses.DraftExpense,
			wantChanged:     false,
			wantErr:         false,
		},
		{
			name: "amountChanged=true, single item with participant -> ready",
			args: args{
				expense: expenses.GroupExpense{
					Status:    expenses.DraftExpense,
					FeesTotal: decimal.NewFromInt(1000),
					Items: []expenses.ExpenseItem{
						{
							Amount:   decimal.NewFromInt(10000),
							Quantity: 2,
							Participants: []expenses.ItemParticipant{
								{ProfileID: uuid.New()},
							},
						},
					},
				},
				amountChanged: true,
			},
			wantItemsTotal:  decimal.NewFromInt(20000),
			wantTotalAmount: decimal.NewFromInt(21000),
			wantStatus:      expenses.ReadyExpense,
			wantChanged:     true,
			wantErr:         false,
		},
		{
			name: "item without participants keeps draft status",
			args: args{
				expense: expenses.GroupExpense{
					Status: expenses.DraftExpense,
					Items: []expenses.ExpenseItem{
						{
							Amount:       decimal.NewFromInt(5000),
							Quantity:     1,
							Participants: nil,
						},
					},
				},
				amountChanged: true,
			},
			wantItemsTotal:  decimal.NewFromInt(5000),
			wantTotalAmount: decimal.NewFromInt(5000),
			wantStatus:      expenses.DraftExpense,
			wantChanged:     true,
			wantErr:         false,
		},
		{
			name: "status changes even when amountChanged=false",
			args: args{
				expense: expenses.GroupExpense{
					Status: expenses.DraftExpense,
					Items: []expenses.ExpenseItem{
						{
							Amount:   decimal.NewFromInt(1000),
							Quantity: 1,
							Participants: []expenses.ItemParticipant{
								{ProfileID: uuid.New()},
							},
						},
					},
				},
				amountChanged: false,
			},
			wantItemsTotal:  decimal.Zero, // not recalculated
			wantTotalAmount: decimal.Zero,
			wantStatus:      expenses.ReadyExpense,
			wantChanged:     true,
			wantErr:         false,
		},
		{
			name: "zero total amount is allowed",
			args: args{
				expense: expenses.GroupExpense{
					Status: expenses.DraftExpense,
					Items: []expenses.ExpenseItem{
						{
							Amount:   decimal.Zero,
							Quantity: 1,
							Participants: []expenses.ItemParticipant{
								{ProfileID: uuid.New()},
							},
						},
					},
				},
				amountChanged: true,
			},
			wantItemsTotal:  decimal.Zero,
			wantTotalAmount: decimal.Zero,
			wantStatus:      expenses.ReadyExpense,
			wantChanged:     true,
			wantErr:         false,
		},
		{
			name: "negative total amount returns error",
			args: args{
				expense: expenses.GroupExpense{
					Status: expenses.DraftExpense,
					Items: []expenses.ExpenseItem{
						{
							Amount:   decimal.NewFromInt(-10),
							Quantity: 1,
							Participants: []expenses.ItemParticipant{
								{ProfileID: uuid.New()},
							},
						},
					},
				},
				amountChanged: true,
			},
			wantItemsTotal:  decimal.Zero,
			wantTotalAmount: decimal.Zero,
			wantStatus:      expenses.DraftExpense,
			wantChanged:     false,
			wantErr:         true,
		},
		{
			name: "multiple items aggregated correctly",
			args: args{
				expense: expenses.GroupExpense{
					Status:    expenses.DraftExpense,
					FeesTotal: decimal.NewFromInt(500),
					Items: []expenses.ExpenseItem{
						{
							Amount:   decimal.NewFromInt(1000),
							Quantity: 2,
							Participants: []expenses.ItemParticipant{
								{ProfileID: uuid.New()},
							},
						},
						{
							Amount:   decimal.NewFromInt(3000),
							Quantity: 1,
							Participants: []expenses.ItemParticipant{
								{ProfileID: uuid.New()},
							},
						},
					},
				},
				amountChanged: true,
			},
			wantItemsTotal:  decimal.NewFromInt(5000),
			wantTotalAmount: decimal.NewFromInt(5500),
			wantStatus:      expenses.ReadyExpense,
			wantChanged:     true,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := svc.RecalculateExpense(tt.args.expense, tt.args.amountChanged)

			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}

			assert.Equal(t, tt.wantChanged, changed)
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.True(t, got.ItemsTotal.Equal(tt.wantItemsTotal))
			assert.True(t, got.TotalAmount.Equal(tt.wantTotalAmount))
		})
	}
}
