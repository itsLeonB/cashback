package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/ezutil/v2"
)

func NewExpenseItemRequestToData(req dto.NewExpenseItemRequest) expenses.ExpenseItem {
	return expenses.ExpenseItem{
		Name:     req.Name,
		Amount:   req.Amount,
		Quantity: req.Quantity,
	}
}

func UpdateExpenseItemRequestToData(req dto.UpdateExpenseItemRequest) expenses.ExpenseItem {
	return expenses.ExpenseItem{
		Name:         req.Name,
		Amount:       req.Amount,
		Quantity:     req.Quantity,
		Participants: ezutil.MapSlice(req.Participants, itemParticipantRequestToData),
	}
}

func itemParticipantRequestToData(req dto.ItemParticipantRequest) expenses.ItemParticipant {
	return expenses.ItemParticipant{
		ProfileID: req.ProfileID,
		Weight:    req.Weight,
	}
}

func ExpenseItemRequestToEntity(request dto.NewExpenseItemRequest) expenses.ExpenseItem {
	return expenses.ExpenseItem{
		GroupExpenseID: request.GroupExpenseID,
		Name:           request.Name,
		Amount:         request.Amount,
		Quantity:       request.Quantity,
	}
}

func PatchExpenseItemWithRequest(expenseItem expenses.ExpenseItem, request dto.UpdateExpenseItemRequest) expenses.ExpenseItem {
	expenseItem.Name = request.Name
	expenseItem.Amount = request.Amount
	expenseItem.Quantity = request.Quantity
	return expenseItem
}

func ItemParticipantRequestToEntity(itemParticipant dto.ItemParticipantRequest) expenses.ItemParticipant {
	return expenses.ItemParticipant{
		ProfileID: itemParticipant.ProfileID,
		Weight:    itemParticipant.Weight,
	}
}

func getExpenseItemSimpleMapper(userProfileID uuid.UUID) func(item expenses.ExpenseItem) dto.ExpenseItemResponse {
	return func(item expenses.ExpenseItem) dto.ExpenseItemResponse {
		return ExpenseItemToResponse(item, userProfileID)
	}
}

func ExpenseItemToResponse(item expenses.ExpenseItem, userProfileID uuid.UUID) dto.ExpenseItemResponse {
	return dto.ExpenseItemResponse{
		BaseDTO:        BaseToDTO(item.BaseEntity),
		GroupExpenseID: item.GroupExpenseID,
		Name:           item.Name,
		Amount:         item.Amount,
		Quantity:       item.Quantity,
		Participants:   ezutil.MapSlice(item.Participants, getItemParticipantSimpleMapper(userProfileID)),
	}
}

func getItemParticipantSimpleMapper(userProfileID uuid.UUID) func(itemParticipant expenses.ItemParticipant) dto.ItemParticipantResponse {
	return func(itemParticipant expenses.ItemParticipant) dto.ItemParticipantResponse {
		return itemParticipantToResponse(itemParticipant, userProfileID)
	}
}

func itemParticipantToResponse(itemParticipant expenses.ItemParticipant, userProfileID uuid.UUID) dto.ItemParticipantResponse {
	return dto.ItemParticipantResponse{
		Profile:         ProfileToSimple(itemParticipant.Profile, userProfileID),
		Weight:          itemParticipant.Weight,
		AllocatedAmount: itemParticipant.AllocatedAmount,
	}
}
