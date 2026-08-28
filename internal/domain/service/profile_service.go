package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	monetizationSvc "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type profileServiceImpl struct {
	transactor                crud.Transactor
	profileRepo               repository.ProfileRepository
	userRepo                  crud.Repository[users.User]
	friendshipRepo            repository.FriendshipRepository
	friendshipRequestRepo     repository.FriendshipRequestRepository
	debtTransactionRepo       repository.DebtTransactionRepository
	profileTransferMethodRepo repository.ProfileTransferMethodRepository
	groupExpenseRepo          repository.GroupExpenseRepository
	expenseItemRepo           repository.ExpenseItemRepository
	otherFeeRepo              repository.OtherFeeRepository
	notificationRepo          repository.NotificationRepository
	pushSubscriptionRepo      repository.PushSubscriptionRepository
	subscriptionSvc           monetizationSvc.SubscriptionService
	subscriptionLimitSvc      SubscriptionLimitService
	friendshipBalanceService  FriendshipBalanceService
}

func NewProfileService(
	transactor crud.Transactor,
	profileRepo repository.ProfileRepository,
	userRepo crud.Repository[users.User],
	friendshipRepo repository.FriendshipRepository,
	friendshipRequestRepo repository.FriendshipRequestRepository,
	debtTransactionRepo repository.DebtTransactionRepository,
	profileTransferMethodRepo repository.ProfileTransferMethodRepository,
	groupExpenseRepo repository.GroupExpenseRepository,
	expenseItemRepo repository.ExpenseItemRepository,
	otherFeeRepo repository.OtherFeeRepository,
	notificationRepo repository.NotificationRepository,
	pushSubscriptionRepo repository.PushSubscriptionRepository,
	subscriptionSvc monetizationSvc.SubscriptionService,
	subscriptionLimitSvc SubscriptionLimitService,
	friendshipBalanceService FriendshipBalanceService,
) ProfileService {
	return &profileServiceImpl{
		transactor,
		profileRepo,
		userRepo,
		friendshipRepo,
		friendshipRequestRepo,
		debtTransactionRepo,
		profileTransferMethodRepo,
		groupExpenseRepo,
		expenseItemRepo,
		otherFeeRepo,
		notificationRepo,
		pushSubscriptionRepo,
		subscriptionSvc,
		subscriptionLimitSvc,
		friendshipBalanceService,
	}
}

func (ps *profileServiceImpl) Create(ctx context.Context, request dto.NewProfileRequest) (dto.ProfileResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.Create")
	defer span.End()

	var response dto.ProfileResponse

	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		if request.UserID != uuid.Nil {
			spec := crud.Specification[users.UserProfile]{}
			spec.Model.UserID = uuid.NullUUID{UUID: request.UserID, Valid: true}
			existing, err := ps.profileRepo.FindFirst(ctx, spec)
			if err != nil {
				return err
			}
			if !existing.IsZero() {
				response = mapper.ProfileToResponse(existing, "", dto.SubscriptionResponse{})
				return nil
			}
		}

		newProfile := users.UserProfile{
			UserID: uuid.NullUUID{
				UUID:  request.UserID,
				Valid: request.UserID != uuid.Nil,
			},
			Name:         request.Name,
			Avatar:       request.Avatar,
			HomeCurrency: request.HomeCurrency,
		}

		insertedProfile, err := ps.profileRepo.Insert(ctx, newProfile)
		if err != nil {
			return err
		}

		if request.GenerateSlug {
			insertedProfile.Slug = sql.NullString{String: util.GenerateSlug(request.Name, insertedProfile.ID), Valid: true}
			insertedProfile, err = ps.profileRepo.Update(ctx, insertedProfile)
			if err != nil {
				return err
			}
		}

		if request.UserID != uuid.Nil {
			if err = ps.subscriptionSvc.AttachDefaultSubscription(ctx, insertedProfile.ID); err != nil {
				return err
			}
		}

		response = mapper.ProfileToResponse(insertedProfile, "", dto.SubscriptionResponse{})

		return nil
	})

	return response, err
}

func (ps *profileServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (dto.ProfileResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.GetByID")
	defer span.End()

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

	subs, err := ps.subscriptionLimitSvc.GetCurrent(ctx, id)
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	return mapper.ProfileToResponse(profile, email, subs), nil
}

func (ps *profileServiceImpl) GetAll(ctx context.Context) ([]dto.ProfileResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.GetAll")
	defer span.End()

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

		responses = append(responses, mapper.ProfileToResponse(profile, email, dto.SubscriptionResponse{}))
	}

	return responses, nil
}

func (ps *profileServiceImpl) GetAllReal(ctx context.Context) ([]dto.ProfileResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.GetAllReal")
	defer span.End()

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

		responses = append(responses, mapper.ProfileToResponse(profile, user.Email, dto.SubscriptionResponse{}))
	}

	return responses, nil
}

func (ps *profileServiceImpl) GetEntityByID(ctx context.Context, id uuid.UUID) (users.UserProfile, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.GetEntityByID")
	defer span.End()

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

func (ps *profileServiceImpl) GetProfileIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.GetProfileIDByUserID")
	defer span.End()

	spec := crud.Specification[users.UserProfile]{}
	spec.Model.UserID = uuid.NullUUID{UUID: userID, Valid: true}
	profile, err := ps.profileRepo.FindFirst(ctx, spec)
	if err != nil {
		return uuid.Nil, err
	}
	if profile.IsZero() {
		return uuid.Nil, ungerr.NotFoundError(fmt.Sprintf("profile for user %s not found", userID))
	}
	return profile.ID, nil
}

func (ps *profileServiceImpl) Update(ctx context.Context, req dto.UpdateProfileRequest) (dto.ProfileResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.Update")
	defer span.End()

	var response dto.ProfileResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[users.UserProfile]{}
		spec.Model.ID = req.ID
		spec.ForUpdate = true
		profile, err := ps.profileRepo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if profile.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("profile ID: %s is not found", req.ID))
		}

		if req.Name != "" {
			profile.Name = req.Name
		}

		if req.HomeCurrency != "" {
			profile.HomeCurrency = req.HomeCurrency
		}

		if !profile.OnboardedAt.Valid {
			profile.OnboardedAt.Time = time.Now()
			profile.OnboardedAt.Valid = true
		}

		updatedProfile, err := ps.profileRepo.Update(ctx, profile)
		if err != nil {
			return err
		}

		response = mapper.ProfileToResponse(updatedProfile, "", dto.SubscriptionResponse{})
		return nil
	})
	return response, err
}

func (ps *profileServiceImpl) Search(ctx context.Context, profileID uuid.UUID, input string) ([]dto.SearchProfileResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.Search")
	defer span.End()

	if util.IsValidEmail(input) {
		profile, err := ps.GetByEmail(ctx, input)
		if err != nil {
			return nil, err
		}
		if profile.ID == profileID {
			return []dto.SearchProfileResponse{}, nil
		}
		return []dto.SearchProfileResponse{{
			ID:     profile.ID,
			Name:   profile.Name,
			Avatar: profile.Avatar,
		}}, nil
	}

	profiles, err := ps.profileRepo.SearchByName(ctx, input, 10)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.SearchProfileResponse, 0, len(profiles))
	for _, profile := range profiles {
		if profile.ID != profileID {
			responses = append(responses, dto.SearchProfileResponse{
				ID:     profile.ID,
				Name:   profile.Name,
				Avatar: profile.Avatar,
			})
		}
	}

	return responses, nil
}

func (ps *profileServiceImpl) GetByEmail(ctx context.Context, email string) (dto.ProfileResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.GetByEmail")
	defer span.End()

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

	return mapper.ProfileToResponse(user.Profile, user.Email, dto.SubscriptionResponse{}), nil
}

// MergeAnonymousProfile physically merges an anonymous placeholder profile into a real,
// registered profile: every row that references anonProfileID is repointed (or merged, for
// tables where the real profile might already have a colliding row) onto realProfileID, then
// the now-unreferenced anonymous profile row is deleted.
//
// ownerProfileID is the profile that vouches for the merge: it must already be friends with
// both the anonymous placeholder (proving it owns/created it) and the real profile (proving
// the two are already connected). For the slug-based auto-merge flow (see auth_hooks.go), the
// caller creates the owner<->real friendship immediately before calling this method, which is
// what satisfies the second check; MergeAnonymousProfile's own friendship repoint step then
// simply drops the now-redundant owner<->anon row instead of duplicating it.
func (ps *profileServiceImpl) MergeAnonymousProfile(ctx context.Context, ownerProfileID, realProfileID, anonProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.MergeAnonymousProfile")
	defer span.End()

	return ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		if ownerProfileID == uuid.Nil || realProfileID == uuid.Nil || anonProfileID == uuid.Nil {
			return ungerr.BadRequestError("ownerProfileID / realProfileID / anonProfileID cannot be nil")
		}

		realProfile, err := ps.GetEntityByID(ctx, realProfileID)
		if err != nil {
			return err
		}
		if !realProfile.IsReal() {
			return ungerr.BadRequestError("realProfileID must belong to a real profile")
		}

		anonProfile, err := ps.GetEntityByID(ctx, anonProfileID)
		if err != nil {
			return err
		}
		if anonProfile.IsReal() {
			return ungerr.BadRequestError("anonProfileID must belong to an anonymous profile")
		}

		if err := ps.checkFriendship(ctx, ownerProfileID, realProfileID, "real"); err != nil {
			return err
		}
		if err := ps.checkFriendship(ctx, ownerProfileID, anonProfileID, "anonymous"); err != nil {
			return err
		}

		if err := ps.friendshipRepo.RepointFriendships(ctx, anonProfileID, realProfileID); err != nil {
			return err
		}
		if err := ps.friendshipRequestRepo.RepointProfile(ctx, anonProfileID, realProfileID); err != nil {
			return err
		}
		if err := ps.profileTransferMethodRepo.RepointProfile(ctx, anonProfileID, realProfileID); err != nil {
			return err
		}
		if err := ps.debtTransactionRepo.RepointProfile(ctx, anonProfileID, realProfileID); err != nil {
			return err
		}
		// Recalculates every friendship_balances row for realProfileID's whole friend graph in
		// one pass, since RepointProfile above can fold a former anon-counterparty's
		// transactions into a pair that already had independent history with realProfileID.
		if err := ps.friendshipBalanceService.RecalculateAllForProfile(ctx, realProfileID); err != nil {
			return err
		}
		if err := ps.groupExpenseRepo.RepointProfile(ctx, anonProfileID, realProfileID); err != nil {
			return err
		}
		if err := ps.expenseItemRepo.RepointParticipants(ctx, anonProfileID, realProfileID); err != nil {
			return err
		}
		if err := ps.otherFeeRepo.RepointParticipants(ctx, anonProfileID, realProfileID); err != nil {
			return err
		}
		if err := ps.notificationRepo.RepointProfile(ctx, anonProfileID, realProfileID); err != nil {
			return err
		}
		if err := ps.pushSubscriptionRepo.RepointProfile(ctx, anonProfileID, realProfileID); err != nil {
			return err
		}

		return ps.profileRepo.Delete(ctx, anonProfile)
	})
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
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.GetByIDs")
	defer span.End()

	profiles, err := ps.profileRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	profileMap := make(map[uuid.UUID]dto.ProfileResponse, len(profiles))
	for _, profile := range profiles {
		profileMap[profile.ID] = mapper.ProfileToResponse(profile, "", dto.SubscriptionResponse{})
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

func (ps *profileServiceImpl) FindBySlug(ctx context.Context, slug string) (users.UserProfile, error) {
	ctx, span := otel.Tracer.Start(ctx, "ProfileService.FindBySlug")
	defer span.End()

	spec := crud.Specification[users.UserProfile]{}
	spec.Model.Slug = sql.NullString{String: slug, Valid: true}
	profile, err := ps.profileRepo.FindFirst(ctx, spec)
	if err != nil {
		return users.UserProfile{}, err
	}
	if profile.IsZero() {
		return users.UserProfile{}, ungerr.NotFoundError("profile not found")
	}
	return profile, nil
}
