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
	"github.com/itsLeonB/cashback/internal/domain/service/expense"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type expenseItemServiceImpl struct {
	transactor            crud.Transactor
	expenseItemRepository repository.ExpenseItemRepository
	groupExpenseSvc       GroupExpenseService
	allocationSvc         expense.AllocationService
}

func NewExpenseItemService(
	transactor crud.Transactor,
	expenseItemRepository repository.ExpenseItemRepository,
	groupExpenseSvc GroupExpenseService,
) ExpenseItemService {
	return &expenseItemServiceImpl{
		transactor,
		expenseItemRepository,
		groupExpenseSvc,
		expense.NewAllocationService(),
	}
}

func (ges *expenseItemServiceImpl) Add(ctx context.Context, req dto.NewExpenseItemRequest) error {
	if req.Amount.IsZero() {
		return ungerr.UnprocessableEntityError(appconstant.ErrAmountZero)
	}

	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expenseItem := mapper.ExpenseItemRequestToEntity(req)
		_, err := ges.expenseItemRepository.Insert(ctx, expenseItem)
		if err != nil {
			return err
		}

		return ges.groupExpenseSvc.Recalculate(ctx, req.UserProfileID, req.GroupExpenseID, true)
	})
}

func (ges *expenseItemServiceImpl) Update(ctx context.Context, req dto.UpdateExpenseItemRequest) error {
	if req.Amount.IsZero() {
		return ungerr.UnprocessableEntityError(appconstant.ErrAmountZero)
	}

	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// 1. Get expense item with validation
		expenseItem, err := ges.getItemForUpdate(ctx, req.ID, req.GroupExpenseID)
		if err != nil {
			return err
		}

		// 2. Patch and update expense item
		patchedExpenseItem := mapper.PatchExpenseItemWithRequest(expenseItem, req)
		updatedExpenseItem, err := ges.expenseItemRepository.Update(ctx, patchedExpenseItem)
		if err != nil {
			return err
		}

		amountChanged := expenseItem.TotalAmount().Compare(updatedExpenseItem.TotalAmount()) != 0

		// 3. Handle amount change - update allocations if amount changed
		if amountChanged {
			if _, err = ges.allocateAndSyncParticipants(ctx, updatedExpenseItem); err != nil {
				return err
			}
		}

		return ges.groupExpenseSvc.Recalculate(ctx, req.UserProfileID, req.GroupExpenseID, amountChanged)
	})
}

func (ges *expenseItemServiceImpl) Remove(ctx context.Context, groupExpenseID, expenseItemID, userProfileID uuid.UUID) error {
	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expenseItem, err := ges.getItemForUpdate(ctx, expenseItemID, groupExpenseID)
		if err != nil {
			return err
		}

		if err = ges.expenseItemRepository.Delete(ctx, expenseItem); err != nil {
			return err
		}

		return ges.groupExpenseSvc.Recalculate(ctx, userProfileID, groupExpenseID, true)
	})
}

func (ges *expenseItemServiceImpl) SyncParticipants(ctx context.Context, req dto.SyncItemParticipantsRequest) error {
	return ges.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		expenseItem, err := ges.getItemForUpdate(ctx, req.ID, req.GroupExpenseID)
		if err != nil {
			return err
		}

		expenseItem.Participants = ezutil.MapSlice(req.Participants, mapper.ItemParticipantRequestToEntity)
		expenseItem, err = ges.allocateAndSyncParticipants(ctx, expenseItem)
		if err != nil {
			return err
		}

		return ges.groupExpenseSvc.Recalculate(ctx, req.ProfileID, req.GroupExpenseID, false)
	})
}

func (ges *expenseItemServiceImpl) allocateAndSyncParticipants(ctx context.Context, expenseItem expenses.ExpenseItem) (expenses.ExpenseItem, error) {
	allocatedParticipants, err := ges.allocationSvc.AllocateAmounts(expenseItem.TotalAmount(), expenseItem.Participants)
	if err != nil {
		return expenses.ExpenseItem{}, err
	}

	if err := ges.expenseItemRepository.SyncParticipants(ctx, expenseItem.ID, allocatedParticipants); err != nil {
		return expenses.ExpenseItem{}, err
	}

	expenseItem.Participants = allocatedParticipants
	return expenseItem, nil
}

func (ges *expenseItemServiceImpl) getItemForUpdate(ctx context.Context, expenseItemID, groupExpenseID uuid.UUID) (expenses.ExpenseItem, error) {
	spec := crud.Specification[expenses.ExpenseItem]{}
	spec.Model.ID = expenseItemID
	spec.Model.GroupExpenseID = groupExpenseID
	spec.ForUpdate = true
	spec.PreloadRelations = []string{"Participants"}

	expenseItem, err := ges.expenseItemRepository.FindFirst(ctx, spec)
	if err != nil {
		return expenses.ExpenseItem{}, err
	}
	if expenseItem.IsZero() {
		return expenses.ExpenseItem{}, ungerr.NotFoundError(fmt.Sprintf("expense item with ID %s is not found", spec.Model.ID))
	}

	return expenseItem, nil
}
