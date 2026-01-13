package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type profileTransferMethodService struct {
	profileSvc                ProfileService
	profileTransferMethodRepo crud.Repository[debts.ProfileTransferMethod]
	transferMethodSvc         TransferMethodService
	friendshipSvc             FriendshipService
}

func NewProfileTransferMethodService(
	profileSvc ProfileService,
	profileTransferMethodRepo crud.Repository[debts.ProfileTransferMethod],
	transferMethodSvc TransferMethodService,
	friendshipSvc FriendshipService,
) *profileTransferMethodService {
	return &profileTransferMethodService{
		profileSvc,
		profileTransferMethodRepo,
		transferMethodSvc,
		friendshipSvc,
	}
}

func (ptm *profileTransferMethodService) Add(ctx context.Context, req dto.NewProfileTransferMethodRequest) error {
	if _, err := ptm.profileSvc.GetEntityByID(ctx, req.ProfileID); err != nil {
		return err
	}

	method, err := ptm.transferMethodSvc.GetByID(ctx, req.TransferMethodID)
	if err != nil {
		return err
	}

	if !method.ParentID.Valid {
		return ungerr.UnprocessableEntityError("cannot add parent transfer method to profile")
	}

	newProfileMethod := debts.ProfileTransferMethod{
		ProfileID:        req.ProfileID,
		TransferMethodID: req.TransferMethodID,
		AccountName:      req.AccountName,
		AccountNumber:    req.AccountNumber,
	}

	if _, err := ptm.profileTransferMethodRepo.Insert(ctx, newProfileMethod); err != nil {
		return ungerr.Wrap(err, "error inserting new profile transfer method")
	}
	return nil
}

func (ptm *profileTransferMethodService) GetAllByProfileID(ctx context.Context, profileID uuid.UUID) ([]dto.ProfileTransferMethodResponse, error) {
	if _, err := ptm.profileSvc.GetEntityByID(ctx, profileID); err != nil {
		return nil, err
	}

	return ptm.getByProfileID(ctx, profileID)
}

func (ptm *profileTransferMethodService) GetAllByFriendProfileID(ctx context.Context, userProfileID, friendProfileID uuid.UUID) ([]dto.ProfileTransferMethodResponse, error) {
	if _, err := ptm.profileSvc.GetByIDs(ctx, []uuid.UUID{userProfileID, friendProfileID}); err != nil {
		return nil, err
	}

	isFriends, _, err := ptm.friendshipSvc.IsFriends(ctx, userProfileID, friendProfileID)
	if err != nil {
		return nil, err
	}
	if !isFriends {
		return nil, ungerr.ForbiddenError("users are not friends")
	}

	return ptm.getByProfileID(ctx, friendProfileID)
}

func (ptm *profileTransferMethodService) getByProfileID(ctx context.Context, profileID uuid.UUID) ([]dto.ProfileTransferMethodResponse, error) {
	spec := crud.Specification[debts.ProfileTransferMethod]{}
	spec.Model.ProfileID = profileID
	spec.PreloadRelations = []string{"Method"}
	methods, err := ptm.profileTransferMethodRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(methods, mapper.ProfileTransferMethodPopulator(ptm.transferMethodSvc.SignedURLPopulator(ctx))), nil
}
