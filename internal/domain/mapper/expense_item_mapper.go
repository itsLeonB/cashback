package mapper

import (
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
		Share:     req.Share,
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
		Share:     itemParticipant.Share,
	}
}
