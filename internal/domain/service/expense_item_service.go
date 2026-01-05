package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type expenseItemServiceImpl struct {
	transactor             crud.Transactor
	groupExpenseRepository repository.GroupExpenseRepository
	expenseItemRepository  repository.ExpenseItemRepository
	groupExpenseSvc        GroupExpenseService
}

func NewExpenseItemService(
	transactor crud.Transactor,
	groupExpenseRepository repository.GroupExpenseRepository,
	expenseItemRepository repository.ExpenseItemRepository,
	groupExpenseSvc GroupExpenseService,
) ExpenseItemService {
	return &expenseItemServiceImpl{
		transactor,
		groupExpenseRepository,
		expenseItemRepository,
		groupExpenseSvc,
	}
}

func (ges *expenseItemServiceImpl) Add(ctx context.Context, req dto.NewExpenseItemRequest) (dto.ExpenseItemResponse, error) {
	var response dto.ExpenseItemResponse

	if !req.Amount.IsPositive() {
		return dto.ExpenseItemResponse{}, ungerr.UnprocessableEntityError(appconstant.ErrNonPositiveAmount)
	}

	err := ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		groupExpense, err := ges.groupExpenseSvc.GetUnconfirmedGroupExpenseForUpdate(ctx, req.UserProfileID, req.GroupExpenseID)
		if err != nil {
			return err
		}

		expenseItem := mapper.ExpenseItemRequestToEntity(req)

		itemTotalAmount := expenseItem.TotalAmount()
		groupExpense.TotalAmount = groupExpense.TotalAmount.Add(itemTotalAmount)
		groupExpense.ItemsTotal = groupExpense.ItemsTotal.Add(itemTotalAmount)
		groupExpense.Status = expenses.DraftExpense
		if _, err = ges.groupExpenseRepository.Update(ctx, groupExpense); err != nil {
			return err
		}

		insertedItem, err := ges.expenseItemRepository.Insert(ctx, expenseItem)
		if err != nil {
			return err
		}

		response = mapper.ExpenseItemToResponse(insertedItem, req.UserProfileID)

		return nil
	})
	return response, err
}

func (ges *expenseItemServiceImpl) Update(ctx context.Context, req dto.UpdateExpenseItemRequest) (dto.ExpenseItemResponse, error) {
	var response dto.ExpenseItemResponse

	if !req.Amount.IsPositive() {
		return dto.ExpenseItemResponse{}, ungerr.UnprocessableEntityError(appconstant.ErrNonPositiveAmount)
	}

	err := ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expenseItem, err := ges.getExpenseItemByIDForUpdate(ctx, req.ID, req.GroupExpenseID)
		if err != nil {
			return err
		}

		if ezutil.CompareUUID(req.GroupExpenseID, expenseItem.GroupExpenseID) != 0 {
			return ungerr.UnprocessableEntityError("mismatched group expense ID")
		}

		groupExpense, err := ges.groupExpenseSvc.GetUnconfirmedGroupExpenseForUpdate(ctx, req.UserProfileID, expenseItem.GroupExpenseID)
		if err != nil {
			return err
		}

		patchedExpenseItem := mapper.PatchExpenseItemWithRequest(expenseItem, req)

		updatedExpenseItem, err := ges.expenseItemRepository.Update(ctx, patchedExpenseItem)
		if err != nil {
			return err
		}

		if len(req.Participants) > 0 {
			updatedParticipants := ezutil.MapSlice(req.Participants, mapper.ItemParticipantRequestToEntity)
			if err := ges.expenseItemRepository.SyncParticipants(ctx, updatedExpenseItem.ID, updatedParticipants); err != nil {
				return err
			}

			updatedExpenseItem.Participants = updatedParticipants
		}

		newItemSet := []expenses.ExpenseItem{}
		for _, item := range groupExpense.Items {
			if item.ID == updatedExpenseItem.ID {
				newItemSet = append(newItemSet, updatedExpenseItem)
			} else {
				newItemSet = append(newItemSet, item)
			}
		}

		if isReadyExpense(newItemSet) {
			groupExpense.Status = expenses.ReadyExpense
		} else {
			groupExpense.Status = expenses.DraftExpense
		}

		oldAmount := expenseItem.TotalAmount()
		newAmount := updatedExpenseItem.TotalAmount()

		if oldAmount.Cmp(newAmount) != 0 {
			groupExpense.TotalAmount = groupExpense.TotalAmount.
				Sub(oldAmount).
				Add(newAmount)

			groupExpense.ItemsTotal = groupExpense.ItemsTotal.
				Sub(oldAmount).
				Add(newAmount)

			if _, err := ges.groupExpenseRepository.Update(ctx, groupExpense); err != nil {
				return err
			}
		}

		response = mapper.ExpenseItemToResponse(updatedExpenseItem, req.UserProfileID)

		return nil
	})
	return response, err
}

func isReadyExpense(items []expenses.ExpenseItem) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if len(item.Participants) == 0 {
			return false
		}
	}
	return true
}

func (ges *expenseItemServiceImpl) getExpenseItemByIDForUpdate(ctx context.Context, expenseItemID, groupExpenseID uuid.UUID) (expenses.ExpenseItem, error) {
	spec := crud.Specification[expenses.ExpenseItem]{}
	spec.Model.ID = expenseItemID
	spec.Model.GroupExpenseID = groupExpenseID
	spec.ForUpdate = true
	spec.PreloadRelations = []string{"Participants"}

	expenseItem, err := ges.getExpenseItemBySpec(ctx, spec)
	if err != nil {
		return expenses.ExpenseItem{}, err
	}

	return expenseItem, nil
}

func (ges *expenseItemServiceImpl) getExpenseItemBySpec(ctx context.Context, spec crud.Specification[expenses.ExpenseItem]) (expenses.ExpenseItem, error) {
	expenseItem, err := ges.expenseItemRepository.FindFirst(ctx, spec)
	if err != nil {
		return expenses.ExpenseItem{}, err
	}
	if expenseItem.IsZero() {
		return expenses.ExpenseItem{}, ungerr.NotFoundError(fmt.Sprintf("expense item with ID %s is not found", spec.Model.ID))
	}
	return expenseItem, nil
}

func (ges *expenseItemServiceImpl) Remove(ctx context.Context, groupExpenseID, expenseItemID, userProfileID uuid.UUID) error {
	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		groupExpense, err := ges.groupExpenseSvc.GetUnconfirmedGroupExpenseForUpdate(ctx, userProfileID, groupExpenseID)
		if err != nil {
			return err
		}

		expenseItem, err := ges.getExpenseItemByIDForUpdate(ctx, expenseItemID, groupExpenseID)
		if err != nil {
			return err
		}

		itemAmount := expenseItem.TotalAmount()
		groupExpense.TotalAmount = groupExpense.TotalAmount.Sub(itemAmount)
		groupExpense.ItemsTotal = groupExpense.ItemsTotal.Sub(itemAmount)

		if len(groupExpense.Items) <= 1 {
			// Removing an item results in empty items
			groupExpense.Status = expenses.DraftExpense
		}

		if _, err = ges.groupExpenseRepository.Update(ctx, groupExpense); err != nil {
			return err
		}

		if err = ges.expenseItemRepository.Delete(ctx, expenseItem); err != nil {
			return err
		}

		return nil
	})
}

func (ges *expenseItemServiceImpl) SyncParticipants(ctx context.Context, req dto.SyncItemParticipantsRequest) error {
	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expenseItem, err := ges.getExpenseItemByIDForUpdate(ctx, req.ID, req.GroupExpenseID)
		if err != nil {
			return err
		}

		groupExpense, err := ges.groupExpenseSvc.GetUnconfirmedGroupExpenseForUpdate(ctx, req.ProfileID, expenseItem.GroupExpenseID)
		if err != nil {
			return err
		}

		updatedParticipants := ezutil.MapSlice(req.Participants, mapper.ItemParticipantRequestToEntity)
		expenseItem.Participants = updatedParticipants

		newItemSet := []expenses.ExpenseItem{}
		for _, item := range groupExpense.Items {
			if item.ID == expenseItem.ID {
				newItemSet = append(newItemSet, expenseItem)
			} else {
				newItemSet = append(newItemSet, item)
			}
		}

		if isReadyExpense(newItemSet) {
			groupExpense.Status = expenses.ReadyExpense
		} else {
			groupExpense.Status = expenses.DraftExpense
		}

		if _, err := ges.groupExpenseRepository.Update(ctx, groupExpense); err != nil {
			return err
		}

		return ges.expenseItemRepository.SyncParticipants(ctx, expenseItem.ID, updatedParticipants)
	})
}
