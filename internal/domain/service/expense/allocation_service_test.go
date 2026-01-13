package expense_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/service/expense"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAllocationService_EqualSplit(t *testing.T) {
	service := expense.NewAllocationService()
	totalAmount := decimal.NewFromFloat(100.00)

	participants := []expenses.ItemParticipant{
		{ProfileID: uuid.New(), Weight: 0},
		{ProfileID: uuid.New(), Weight: 0},
		{ProfileID: uuid.New(), Weight: 0},
	}

	result, err := service.AllocateAmounts(totalAmount, participants)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Check that all participants have weight = 1 (equal split)
	for _, p := range result {
		assert.Equal(t, 1, p.Weight)
	}

	// Check that total allocated amount equals total amount
	totalAllocated := decimal.Zero
	for _, p := range result {
		totalAllocated = totalAllocated.Add(p.AllocatedAmount)
	}
	assert.True(t, totalAllocated.Equal(totalAmount))
}

func TestAllocationService_WeightedSplit(t *testing.T) {
	service := expense.NewAllocationService()
	totalAmount := decimal.NewFromFloat(100.00)

	participants := []expenses.ItemParticipant{
		{ProfileID: uuid.New(), Weight: 2},
		{ProfileID: uuid.New(), Weight: 3},
	}

	result, err := service.AllocateAmounts(totalAmount, participants)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Check weights are preserved
	assert.Equal(t, 2, result[0].Weight)
	assert.Equal(t, 3, result[1].Weight)

	// Check allocation: 2/5 * 100 = 40, 3/5 * 100 = 60
	assert.True(t, result[0].AllocatedAmount.Equal(decimal.NewFromFloat(40.00)))
	assert.True(t, result[1].AllocatedAmount.Equal(decimal.NewFromFloat(60.00)))

	// Check that total allocated amount equals total amount
	totalAllocated := result[0].AllocatedAmount.Add(result[1].AllocatedAmount)
	assert.True(t, totalAllocated.Equal(totalAmount))
}

func TestAllocationService_MixedWeights_ShouldError(t *testing.T) {
	service := expense.NewAllocationService()
	totalAmount := decimal.NewFromFloat(100.00)

	participants := []expenses.ItemParticipant{
		{ProfileID: uuid.New(), Weight: 2},
		{ProfileID: uuid.New(), Weight: 0},
	}

	_, err := service.AllocateAmounts(totalAmount, participants)

	assert.Error(t, err)
}

func TestAllocationService_NegativeWeight_ShouldError(t *testing.T) {
	service := expense.NewAllocationService()
	totalAmount := decimal.NewFromFloat(100.00)

	participants := []expenses.ItemParticipant{
		{ProfileID: uuid.New(), Weight: -1},
	}

	_, err := service.AllocateAmounts(totalAmount, participants)

	assert.Error(t, err)
}

func TestAllocationService_EmptyParticipants_ShouldError(t *testing.T) {
	service := expense.NewAllocationService()
	totalAmount := decimal.NewFromFloat(100.00)

	participants := []expenses.ItemParticipant{}

	_, err := service.AllocateAmounts(totalAmount, participants)

	assert.Error(t, err)
}

func TestAllocationService_RoundingRemainder(t *testing.T) {
	service := expense.NewAllocationService()
	totalAmount := decimal.NewFromFloat(100.01) // Amount that will cause rounding issues

	participants := []expenses.ItemParticipant{
		{ProfileID: uuid.New(), Weight: 1},
		{ProfileID: uuid.New(), Weight: 1},
		{ProfileID: uuid.New(), Weight: 1},
	}

	result, err := service.AllocateAmounts(totalAmount, participants)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Check that total allocated amount equals total amount exactly
	totalAllocated := decimal.Zero
	for _, p := range result {
		totalAllocated = totalAllocated.Add(p.AllocatedAmount)
	}
	assert.True(t, totalAllocated.Equal(totalAmount))
}
