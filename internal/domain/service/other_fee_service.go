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
	"github.com/itsLeonB/cashback/internal/domain/service/fee"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"github.com/shopspring/decimal"
)

type otherFeeServiceImpl struct {
	transactor             crud.Transactor
	groupExpenseRepository repository.GroupExpenseRepository
	feeCalculatorRegistry  map[expenses.FeeCalculationMethod]fee.FeeCalculator
	otherFeeRepository     repository.OtherFeeRepository
	groupExpenseSvc        GroupExpenseService
}

func NewOtherFeeService(
	transactor crud.Transactor,
	groupExpenseRepository repository.GroupExpenseRepository,
	otherFeeRepository repository.OtherFeeRepository,
	groupExpenseSvc GroupExpenseService,
) OtherFeeService {
	return &otherFeeServiceImpl{
		transactor,
		groupExpenseRepository,
		fee.NewFeeCalculatorRegistry(),
		otherFeeRepository,
		groupExpenseSvc,
	}
}

func (ofs *otherFeeServiceImpl) Add(ctx context.Context, req dto.NewOtherFeeRequest) (dto.OtherFeeResponse, error) {
	var response dto.OtherFeeResponse

	if !req.Amount.IsPositive() {
		return dto.OtherFeeResponse{}, ungerr.UnprocessableEntityError(appconstant.ErrNonPositiveAmount)
	}

	err := ofs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		groupExpense, err := ofs.groupExpenseSvc.GetUnconfirmedGroupExpenseForUpdate(ctx, req.UserProfileID, req.GroupExpenseID)
		if err != nil {
			return err
		}

		fee := mapper.OtherFeeRequestToEntity(req)

		groupExpense.TotalAmount = groupExpense.TotalAmount.Add(fee.Amount)
		groupExpense.FeesTotal = groupExpense.FeesTotal.Add(fee.Amount)
		if _, err = ofs.groupExpenseRepository.Update(ctx, groupExpense); err != nil {
			return err
		}

		insertedFee, err := ofs.otherFeeRepository.Insert(ctx, fee)
		if err != nil {
			return err
		}

		response = mapper.OtherFeeToResponse(insertedFee, req.UserProfileID)

		return nil
	})
	return response, err
}

func (ofs *otherFeeServiceImpl) Update(ctx context.Context, req dto.UpdateOtherFeeRequest) (dto.OtherFeeResponse, error) {
	var response dto.OtherFeeResponse

	if req.Amount.Cmp(decimal.Zero) <= 0 {
		return dto.OtherFeeResponse{}, ungerr.UnprocessableEntityError("amount must be more than 0")
	}

	err := ofs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		groupExpense, err := ofs.groupExpenseSvc.GetUnconfirmedGroupExpenseForUpdate(ctx, req.UserProfileID, req.GroupExpenseID)
		if err != nil {
			return err
		}

		spec := crud.Specification[expenses.OtherFee]{}
		spec.Model.ID = req.ID
		spec.Model.GroupExpenseID = req.GroupExpenseID
		spec.ForUpdate = true
		otherFee, err := ofs.otherFeeRepository.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if otherFee.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("other fee with ID: %s is not found", req.ID))
		}

		patchedFee := mapper.PatchOtherFeeWithRequest(otherFee, req)

		updatedFee, err := ofs.otherFeeRepository.Update(ctx, patchedFee)
		if err != nil {
			return err
		}

		if updatedFee.Amount.Cmp(otherFee.Amount) != 0 {
			groupExpense.TotalAmount = groupExpense.TotalAmount.Sub(otherFee.Amount).Add(updatedFee.Amount)
			groupExpense.FeesTotal = groupExpense.FeesTotal.Sub(otherFee.Amount).Add(updatedFee.Amount)
			if _, err = ofs.groupExpenseRepository.Update(ctx, groupExpense); err != nil {
				return err
			}
		}

		response = mapper.OtherFeeToResponse(updatedFee, req.UserProfileID)

		return nil
	})
	return response, err
}

func (ofs *otherFeeServiceImpl) Remove(ctx context.Context, groupExpenseID, otherFeeID, userProfileID uuid.UUID) error {
	return ofs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		groupExpense, err := ofs.groupExpenseSvc.GetUnconfirmedGroupExpenseForUpdate(ctx, userProfileID, groupExpenseID)
		if err != nil {
			return err
		}

		spec := crud.Specification[expenses.OtherFee]{}
		spec.Model.ID = otherFeeID
		spec.Model.GroupExpenseID = groupExpenseID
		spec.ForUpdate = true
		otherFee, err := ofs.otherFeeRepository.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if otherFee.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("other fee with ID: %s is not found", otherFeeID))
		}

		if err = ofs.otherFeeRepository.Delete(ctx, otherFee); err != nil {
			return err
		}

		groupExpense.TotalAmount = groupExpense.TotalAmount.Sub(otherFee.Amount)
		groupExpense.FeesTotal = groupExpense.FeesTotal.Sub(otherFee.Amount)
		if _, err = ofs.groupExpenseRepository.Update(ctx, groupExpense); err != nil {
			return err
		}

		return nil
	})
}

func (ofs *otherFeeServiceImpl) GetCalculationMethods(ctx context.Context) []dto.FeeCalculationMethodInfo {
	feeCalculationMethodInfos := make([]dto.FeeCalculationMethodInfo, 0, len(ofs.feeCalculatorRegistry))
	for _, feeCalculator := range ofs.feeCalculatorRegistry {
		feeCalculationMethodInfos = append(feeCalculationMethodInfos, feeCalculator.GetInfo())
	}

	return feeCalculationMethodInfos
}
