package expense

import (
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
	"github.com/shopspring/decimal"
)

type AllocationService interface {
	AllocateAmounts(totalAmount decimal.Decimal, participants []expenses.ItemParticipant) ([]expenses.ItemParticipant, error)
}

type allocationServiceImpl struct{}

func NewAllocationService() AllocationService {
	return &allocationServiceImpl{}
}

func (a *allocationServiceImpl) AllocateAmounts(totalAmount decimal.Decimal, participants []expenses.ItemParticipant) ([]expenses.ItemParticipant, error) {
	if len(participants) == 0 {
		return nil, ungerr.UnprocessableEntityError("no participants provided")
	}

	// Calculate sum of weights
	weightSum := 0
	hasPositiveWeight := false
	hasZeroWeight := false

	for _, p := range participants {
		if p.Weight > 0 {
			hasPositiveWeight = true
			weightSum = weightSum + p.Weight
		} else if p.Weight == 0 {
			hasZeroWeight = true
		} else {
			return nil, ungerr.UnprocessableEntityError("weight cannot be negative")
		}
	}

	// Validate weight consistency
	if hasPositiveWeight && hasZeroWeight {
		return nil, ungerr.UnprocessableEntityError("mixed weighted and unweighted participants")
	}

	// Determine effective weights
	effectiveWeights := make([]int, len(participants))
	effectiveWeightSum := 0

	if weightSum == 0 {
		// Equal split - assign weight = 1 to all participants
		for i := range participants {
			effectiveWeights[i] = 1
			effectiveWeightSum++
		}
	} else {
		// Use provided weights
		for i, p := range participants {
			if p.Weight <= 0 {
				return nil, ungerr.UnprocessableEntityError("all weights must be positive when any weight is provided")
			}
			effectiveWeights[i] = p.Weight
			effectiveWeightSum = effectiveWeightSum + p.Weight
		}
	}

	// Calculate unit value
	unitValue := totalAmount.Div(decimal.NewFromInt(int64(effectiveWeightSum)))

	// Allocate amounts
	allocatedSum := decimal.Zero
	result := make([]expenses.ItemParticipant, len(participants))

	for i, p := range participants {
		result[i] = p
		result[i].Weight = effectiveWeights[i]
		result[i].AllocatedAmount = decimal.NewFromInt(int64(effectiveWeights[i])).Mul(unitValue).Round(2)
		allocatedSum = allocatedSum.Add(result[i].AllocatedAmount)
	}

	// Handle rounding remainder
	remainder := totalAmount.Sub(allocatedSum)
	if !remainder.IsZero() {
		// Find participant with highest weight, use ProfileID as tiebreaker
		maxWeightIdx := 0
		maxWeight := effectiveWeights[0]
		minProfileID := participants[0].ProfileID

		for i := 1; i < len(participants); i++ {
			if effectiveWeights[i] > maxWeight ||
				(effectiveWeights[i] == maxWeight && ezutil.CompareUUID(participants[i].ProfileID, minProfileID) < 0) {
				maxWeightIdx = i
				maxWeight = effectiveWeights[i]
				minProfileID = participants[i].ProfileID
			}
		}

		result[maxWeightIdx].AllocatedAmount = result[maxWeightIdx].AllocatedAmount.Add(remainder)
	}

	// Final validation
	finalSum := decimal.Zero
	for _, p := range result {
		finalSum = finalSum.Add(p.AllocatedAmount)
	}

	if !finalSum.Equal(totalAmount) {
		return nil, ungerr.InternalServerError()
	}

	return result, nil
}
