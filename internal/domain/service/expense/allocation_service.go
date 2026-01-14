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

	// Create a copy of participants to avoid modifying the input
	result := make([]expenses.ItemParticipant, len(participants))
	copy(result, participants)

	// Calculate and validate weights
	weightSum, err := calculateAndValidateWeights(participants)
	if err != nil {
		return nil, err
	}

	// Determine weights to use
	weights := determineWeights(participants, weightSum)
	weightTotal := sumInts(weights)

	// Calculate unit value and allocate amounts
	unitValue := totalAmount.Div(decimal.NewFromInt(int64(weightTotal)))
	allocatedSum, err := allocateAmounts(result, weights, unitValue)
	if err != nil {
		return nil, err
	}

	// Handle rounding remainder
	remainder := totalAmount.Sub(allocatedSum)
	if !remainder.IsZero() {
		applyRemainder(result, weights, remainder)
	}

	// Validate final sum
	if err := validateFinalSum(result, totalAmount); err != nil {
		return nil, err
	}

	return result, nil
}

// Helper function to determine which weights to use
func determineWeights(participants []expenses.ItemParticipant, weightSum int) []int {
	weights := make([]int, len(participants))

	if weightSum == 0 {
		// Equal split
		for i := range weights {
			weights[i] = 1
		}
	} else {
		// Validate and use provided weights
		for i, p := range participants {
			if p.Weight <= 0 {
				// Error should have been caught by calculateAndValidateWeights
				weights[i] = 1
			} else {
				weights[i] = p.Weight
			}
		}
	}
	return weights
}

// Helper function to allocate amounts based on weights
func allocateAmounts(participants []expenses.ItemParticipant, weights []int, unitValue decimal.Decimal) (decimal.Decimal, error) {
	if len(participants) != len(weights) {
		return decimal.Zero, ungerr.Unknown("participants and weights length mismatch")
	}

	allocatedSum := decimal.Zero
	var weightMultiplier decimal.Decimal

	for i := range participants {
		participants[i].Weight = weights[i]

		// Reuse decimal to avoid allocations
		weightMultiplier = decimal.NewFromInt(int64(weights[i]))
		participants[i].AllocatedAmount = weightMultiplier.Mul(unitValue).Round(2)

		allocatedSum = allocatedSum.Add(participants[i].AllocatedAmount)
	}

	return allocatedSum, nil
}

// Helper function to apply remainder to participant with highest weight
func applyRemainder(participants []expenses.ItemParticipant, weights []int, remainder decimal.Decimal) {
	if len(participants) == 0 || len(weights) == 0 {
		return
	}

	// Find index with highest weight (using ProfileID as tiebreaker)
	maxIdx := 0
	maxWeight := weights[0]
	minProfileID := participants[0].ProfileID

	for i := 1; i < len(weights); i++ {
		currentWeight := weights[i]
		currentProfileID := participants[i].ProfileID

		if currentWeight > maxWeight ||
			(currentWeight == maxWeight && ezutil.CompareUUID(currentProfileID, minProfileID) < 0) {
			maxIdx = i
			maxWeight = currentWeight
			minProfileID = currentProfileID
		}
	}

	participants[maxIdx].AllocatedAmount = participants[maxIdx].AllocatedAmount.Add(remainder)
}

// Helper function to sum integers
func sumInts(nums []int) int {
	total := 0
	for _, num := range nums {
		total += num
	}
	return total
}

func validateFinalSum(result []expenses.ItemParticipant, totalAmount decimal.Decimal) error {
	finalSum := decimal.Zero
	for _, p := range result {
		finalSum = finalSum.Add(p.AllocatedAmount)
	}

	if !finalSum.Equal(totalAmount) {
		return ungerr.Unknownf("fail to allocate amounts, calculated finalSum: %s, totalAmount: %s", finalSum.String(), totalAmount.String())
	}

	return nil
}

func calculateAndValidateWeights(participants []expenses.ItemParticipant) (int, error) {
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
			return 0, ungerr.UnprocessableEntityError("weight cannot be negative")
		}
	}

	// Validate weight consistency
	if hasPositiveWeight && hasZeroWeight {
		return 0, ungerr.UnprocessableEntityError("mixed weighted and unweighted participants")
	}

	return weightSum, nil
}
