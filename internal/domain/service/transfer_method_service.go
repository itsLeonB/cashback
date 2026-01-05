package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type transferMethodServiceImpl struct {
	transferMethodRepository crud.Repository[debts.TransferMethod]
}

func NewTransferMethodService(transferMethodRepository crud.Repository[debts.TransferMethod]) TransferMethodService {
	return &transferMethodServiceImpl{transferMethodRepository}
}

func (tms *transferMethodServiceImpl) GetAll(ctx context.Context) ([]dto.TransferMethodResponse, error) {
	transferMethods, err := tms.transferMethodRepository.FindAll(ctx, crud.Specification[debts.TransferMethod]{})
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(transferMethods, mapper.TransferMethodToResponse), nil
}

func (tms *transferMethodServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (debts.TransferMethod, error) {
	spec := crud.Specification[debts.TransferMethod]{}
	spec.Model.ID = id

	transferMethod, err := tms.transferMethodRepository.FindFirst(ctx, spec)
	if err != nil {
		return debts.TransferMethod{}, err
	}
	if transferMethod.IsZero() {
		return debts.TransferMethod{}, ungerr.NotFoundError(fmt.Sprintf(appconstant.ErrTransferMethodNotFound, id))
	}

	return transferMethod, nil
}

func (tms *transferMethodServiceImpl) GetByName(ctx context.Context, name string) (debts.TransferMethod, error) {
	spec := crud.Specification[debts.TransferMethod]{}
	spec.Model.Name = name

	transferMethod, err := tms.transferMethodRepository.FindFirst(ctx, spec)
	if err != nil {
		return debts.TransferMethod{}, err
	}
	if transferMethod.IsZero() {
		return debts.TransferMethod{}, ungerr.Unknownf("%s transfer method not found", name)
	}

	return transferMethod, nil
}
