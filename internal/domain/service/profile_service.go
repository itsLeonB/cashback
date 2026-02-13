package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	monetizationDto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	monetizationSvc "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type profileServiceImpl struct {
	transactor         crud.Transactor
	profileRepo        repository.ProfileRepository
	userRepo           crud.Repository[users.User]
	friendshipRepo     repository.FriendshipRepository
	relatedProfileRepo crud.Repository[users.RelatedProfile]
	subscriptionSvc    monetizationSvc.SubscriptionService
}

func NewProfileService(
	transactor crud.Transactor,
	profileRepo repository.ProfileRepository,
	userRepo crud.Repository[users.User],
	friendshipRepo repository.FriendshipRepository,
	relatedProfileRepo crud.Repository[users.RelatedProfile],
	subscriptionSvc monetizationSvc.SubscriptionService,
) ProfileService {
	return &profileServiceImpl{
		transactor,
		profileRepo,
		userRepo,
		friendshipRepo,
		relatedProfileRepo,
		subscriptionSvc,
	}
}

func (ps *profileServiceImpl) Create(ctx context.Context, request dto.NewProfileRequest) (dto.ProfileResponse, error) {
	var response dto.ProfileResponse

	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		newProfile := users.UserProfile{
			UserID: uuid.NullUUID{
				UUID:  request.UserID,
				Valid: request.UserID != uuid.Nil,
			},
			Name:   request.Name,
			Avatar: request.Avatar,
		}

		insertedProfile, err := ps.profileRepo.Insert(ctx, newProfile)
		if err != nil {
			return err
		}

		if request.UserID != uuid.Nil {
			if err = ps.subscriptionSvc.AttachDefaultSubscription(ctx, insertedProfile.ID); err != nil {
				return err
			}
		}

		response = mapper.ProfileToResponse(insertedProfile, "", nil, uuid.Nil, monetizationDto.SubscriptionResponse{})

		return nil
	})

	return response, err
}

func (ps *profileServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (dto.ProfileResponse, error) {
	profile, err := ps.GetEntityByID(ctx, id)
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	var email string
	if profile.IsReal() {
		userSpec := crud.Specification[users.User]{}
		userSpec.Model.ID = profile.UserID.UUID
		user, err := ps.userRepo.FindFirst(ctx, userSpec)
		if err != nil {
			return dto.ProfileResponse{}, err
		}
		email = user.Email
	}

	anonProfileIDs, realProfileID, err := ps.getAssociations(ctx, profile)
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	currentSubscription, err := ps.subscriptionSvc.GetCurrentSubscription(ctx, id)
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	return mapper.ProfileToResponse(profile, email, anonProfileIDs, realProfileID, currentSubscription), nil
}

func (ps *profileServiceImpl) GetAll(ctx context.Context) ([]dto.ProfileResponse, error) {
	spec := crud.Specification[users.UserProfile]{}
	profiles, err := ps.profileRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProfileResponse, 0, len(profiles))
	for _, profile := range profiles {
		var email string
		if profile.IsReal() {
			userSpec := crud.Specification[users.User]{}
			userSpec.Model.ID = profile.UserID.UUID
			user, err := ps.userRepo.FindFirst(ctx, userSpec)
			if err != nil {
				logger.Error(err)
				continue
			}
			email = user.Email
		}

		anonProfileIDs, realProfileID, err := ps.getAssociations(ctx, profile)
		if err != nil {
			logger.Error(err)
			continue
		}

		currentSubscription, err := ps.subscriptionSvc.GetCurrentSubscription(ctx, profile.ID)
		if err != nil {
			logger.Error(err)
			continue
		}

		responses = append(responses, mapper.ProfileToResponse(profile, email, anonProfileIDs, realProfileID, currentSubscription))
	}

	return responses, nil
}

func (ps *profileServiceImpl) GetAllReal(ctx context.Context) ([]dto.ProfileResponse, error) {
	profiles, err := ps.profileRepo.FindRealProfiles(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProfileResponse, 0, len(profiles))
	for _, profile := range profiles {
		userSpec := crud.Specification[users.User]{}
		userSpec.Model.ID = profile.UserID.UUID
		user, err := ps.userRepo.FindFirst(ctx, userSpec)
		if err != nil {
			logger.Error(err)
			continue
		}

		anonProfileIDs, realProfileID, err := ps.getAssociations(ctx, profile)
		if err != nil {
			logger.Error(err)
			continue
		}

		currentSubscription, err := ps.subscriptionSvc.GetCurrentSubscription(ctx, profile.ID)
		if err != nil {
			logger.Error(err)
			continue
		}

		responses = append(responses, mapper.ProfileToResponse(profile, user.Email, anonProfileIDs, realProfileID, currentSubscription))
	}

	return responses, nil
}

func (ps *profileServiceImpl) GetAssociatedIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	profile, err := ps.GetEntityByID(ctx, id)
	if err != nil {
		return nil, err
	}

	anonProfileIDs, realProfileID, err := ps.getAssociations(ctx, profile)
	if err != nil {
		return nil, err
	}

	ids := []uuid.UUID{id}
	if profile.UserID.Valid {
		ids = append(ids, anonProfileIDs...)
	} else {
		if realProfileID != uuid.Nil {
			ids = append(ids, realProfileID)
		}
	}

	return ids, nil
}

func (ps *profileServiceImpl) getAssociations(ctx context.Context, profile users.UserProfile) ([]uuid.UUID, uuid.UUID, error) {
	if profile.IsReal() {
		anonProfileIDs, err := ps.getAssociatedProfileIDs(ctx, profile.ID)
		if err != nil {
			return nil, uuid.Nil, err
		}
		return anonProfileIDs, uuid.Nil, nil
	} else {
		profileID, err := ps.GetRealProfileID(ctx, profile.ID)
		if err != nil {
			return nil, uuid.Nil, err
		}
		return nil, profileID, nil
	}
}

func (ps *profileServiceImpl) getAssociatedProfileIDs(ctx context.Context, realProfileID uuid.UUID) ([]uuid.UUID, error) {
	spec := crud.Specification[users.RelatedProfile]{}
	spec.Model.RealProfileID = realProfileID
	relations, err := ps.relatedProfileRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}
	return ezutil.MapSlice(relations, func(r users.RelatedProfile) uuid.UUID { return r.AnonProfileID }), nil
}

func (ps *profileServiceImpl) GetRealProfileID(ctx context.Context, anonProfileID uuid.UUID) (uuid.UUID, error) {
	spec := crud.Specification[users.RelatedProfile]{}
	spec.Model.AnonProfileID = anonProfileID
	relation, err := ps.relatedProfileRepo.FindFirst(ctx, spec)
	return relation.RealProfileID, err
}

func (ps *profileServiceImpl) GetEntityByID(ctx context.Context, id uuid.UUID) (users.UserProfile, error) {
	spec := crud.Specification[users.UserProfile]{}
	spec.Model.ID = id
	profile, err := ps.profileRepo.FindFirst(ctx, spec)
	if err != nil {
		return users.UserProfile{}, err
	}
	if profile.IsZero() {
		return users.UserProfile{}, ungerr.NotFoundError(fmt.Sprintf("profile with ID: %s is not found", id))
	}
	return profile, nil
}

func (ps *profileServiceImpl) Update(ctx context.Context, id uuid.UUID, name string) (dto.ProfileResponse, error) {
	var response dto.ProfileResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[users.UserProfile]{}
		spec.Model.ID = id
		spec.ForUpdate = true
		profile, err := ps.profileRepo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if profile.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("profile ID: %s is not found", id))
		}

		if name != "" {
			profile.Name = name
		}

		updatedProfile, err := ps.profileRepo.Update(ctx, profile)
		if err != nil {
			return err
		}

		response = mapper.ProfileToResponse(updatedProfile, "", nil, uuid.Nil, monetizationDto.SubscriptionResponse{})
		return nil
	})
	return response, err
}

func (ps *profileServiceImpl) Search(ctx context.Context, profileID uuid.UUID, input string) ([]dto.ProfileResponse, error) {
	if util.IsValidEmail(input) {
		profile, err := ps.GetByEmail(ctx, input)
		if err != nil {
			return nil, err
		}
		if profile.ID == profileID {
			return []dto.ProfileResponse{}, nil
		}
		return []dto.ProfileResponse{profile}, nil
	}

	profiles, err := ps.profileRepo.SearchByName(ctx, input, 10)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProfileResponse, 0, len(profiles))
	for _, profile := range profiles {
		if profile.ID != profileID {
			responses = append(responses, mapper.ProfileToResponse(profile, "", nil, uuid.Nil, monetizationDto.SubscriptionResponse{}))
		}
	}

	return responses, nil
}

func (ps *profileServiceImpl) GetByEmail(ctx context.Context, email string) (dto.ProfileResponse, error) {
	userSpec := crud.Specification[users.User]{}
	userSpec.Model.Email = email
	userSpec.PreloadRelations = []string{"Profile"}
	user, err := ps.userRepo.FindFirst(ctx, userSpec)
	if err != nil {
		return dto.ProfileResponse{}, err
	}
	if user.IsZero() || !user.IsVerified() {
		return dto.ProfileResponse{}, ungerr.NotFoundError("user is not found")
	}

	anonProfileIDs, realProfileID, err := ps.getAssociations(ctx, user.Profile)
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	return mapper.ProfileToResponse(user.Profile, user.Email, anonProfileIDs, realProfileID, monetizationDto.SubscriptionResponse{}), nil
}

func (ps *profileServiceImpl) Associate(ctx context.Context, userProfileID, realProfileID, anonProfileID uuid.UUID) error {
	return ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		if realProfileID == uuid.Nil || anonProfileID == uuid.Nil || userProfileID == uuid.Nil {
			return ungerr.BadRequestError("userProfileID / realProfileID / anonProfileID cannot be nil")
		}

		if _, err := ps.GetEntityByID(ctx, realProfileID); err != nil {
			return err
		}
		if _, err := ps.GetEntityByID(ctx, anonProfileID); err != nil {
			return err
		}

		if err := ps.validateAssociation(ctx, userProfileID, realProfileID, anonProfileID); err != nil {
			return err
		}

		newRelated := users.RelatedProfile{
			RealProfileID: realProfileID,
			AnonProfileID: anonProfileID,
		}
		_, err := ps.relatedProfileRepo.Insert(ctx, newRelated)
		return err
	})
}

func (ps *profileServiceImpl) validateAssociation(ctx context.Context, userProfileID, realProfileID, anonProfileID uuid.UUID) error {
	relatedSpec := crud.Specification[users.RelatedProfile]{}
	relatedSpec.Model.AnonProfileID = anonProfileID
	existingRelated, err := ps.relatedProfileRepo.FindFirst(ctx, relatedSpec)
	if err != nil {
		return err
	}
	if !existingRelated.IsZero() {
		return ungerr.ConflictError("anonProfileID is already associated with a real profile")
	}

	if err := ps.checkFriendship(ctx, userProfileID, realProfileID, "real"); err != nil {
		return err
	}
	if err := ps.checkFriendship(ctx, userProfileID, anonProfileID, "anonymous"); err != nil {
		return err
	}
	return nil
}

func (ps *profileServiceImpl) checkFriendship(ctx context.Context, userProfileID, friendProfileID uuid.UUID, typeStr string) error {
	f, err := ps.friendshipRepo.FindByProfileIDs(ctx, userProfileID, friendProfileID)
	if err != nil {
		return err
	}
	if f.IsZero() {
		return ungerr.ForbiddenError(fmt.Sprintf("user is not friends with the %s profile", typeStr))
	}
	return nil
}

func (ps *profileServiceImpl) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]dto.ProfileResponse, error) {
	profiles, err := ps.profileRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	profileMap := make(map[uuid.UUID]dto.ProfileResponse, len(profiles))
	for _, profile := range profiles {
		profileMap[profile.ID] = mapper.ProfileToResponse(profile, "", nil, uuid.Nil, monetizationDto.SubscriptionResponse{})
	}

	// ensure all requested IDs exist
	var notFoundIDs []uuid.UUID
	for _, id := range ids {
		if _, ok := profileMap[id]; !ok {
			notFoundIDs = append(notFoundIDs, id)
		}
	}

	if len(notFoundIDs) > 0 {
		return nil, ungerr.NotFoundError(fmt.Sprintf("profiles not found: %v", notFoundIDs))
	}

	return profileMap, nil
}
