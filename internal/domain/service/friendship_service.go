package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type friendshipServiceImpl struct {
	transactor           crud.Transactor
	friendshipRepository repository.FriendshipRepository
	profileService       ProfileService
}

func NewFriendshipService(
	transactor crud.Transactor,
	friendshipRepository repository.FriendshipRepository,
	profileService ProfileService,
) FriendshipService {
	return &friendshipServiceImpl{
		transactor,
		friendshipRepository,
		profileService,
	}
}

func (fs *friendshipServiceImpl) CreateAnonymous(ctx context.Context, req dto.NewAnonymousFriendshipRequest) (dto.FriendshipResponse, error) {
	var response dto.FriendshipResponse
	err := fs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		profile, err := fs.profileService.GetByID(ctx, req.ProfileID)
		if err != nil {
			return err
		}

		if err = fs.validateExistingAnonymousFriendship(ctx, profile.ID, req.Name); err != nil {
			return err
		}

		response, err = fs.insertAnonymousFriendship(ctx, profile, req.Name)
		if err != nil {
			return err
		}

		return nil
	})
	return response, err
}

func (fs *friendshipServiceImpl) validateExistingAnonymousFriendship(
	ctx context.Context,
	userProfileID uuid.UUID,
	friendName string,
) error {
	friendshipSpec := users.FriendshipSpecification{}
	friendshipSpec.Model.ProfileID1 = userProfileID
	friendshipSpec.Name = friendName
	friendshipSpec.Model.Type = users.Anonymous

	existingFriendship, err := fs.friendshipRepository.FindFirstBySpec(ctx, friendshipSpec)
	if err != nil {
		return err
	}
	if !existingFriendship.IsZero() {
		return ungerr.ConflictError(fmt.Sprintf("anonymous friend named %s already exists", friendName))
	}

	return nil
}

func (fs *friendshipServiceImpl) insertAnonymousFriendship(
	ctx context.Context,
	userProfile dto.ProfileResponse,
	friendName string,
) (dto.FriendshipResponse, error) {
	newProfile := dto.NewProfileRequest{Name: friendName}

	insertedProfile, err := fs.profileService.Create(ctx, newProfile)
	if err != nil {
		return dto.FriendshipResponse{}, err
	}

	newFriendship, err := mapper.OrderProfilesToFriendship(userProfile, insertedProfile)
	if err != nil {
		return dto.FriendshipResponse{}, err
	}

	newFriendship.Type = users.Anonymous

	insertedFriendship, err := fs.friendshipRepository.Insert(ctx, newFriendship)
	if err != nil {
		return dto.FriendshipResponse{}, err
	}

	return mapper.FriendshipToResponse(userProfile.ID, insertedFriendship)
}

func (fs *friendshipServiceImpl) GetAll(ctx context.Context, profileID uuid.UUID) ([]dto.FriendshipResponse, error) {
	profile, err := fs.profileService.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	spec := users.FriendshipSpecification{}
	spec.Model.ProfileID1 = profile.ID
	spec.PreloadRelations = []string{"Profile1", "Profile2"}

	friendships, err := fs.friendshipRepository.FindAllBySpec(ctx, spec)
	if err != nil {
		return nil, err
	}

	validFriendships := make([]users.Friendship, 0, len(friendships))
	for _, friendship := range friendships {
		_, friendProfile, err := mapper.SelectProfiles(profileID, friendship)
		if err != nil {
			return nil, err
		}

		if friendProfile.IsReal() {
			validFriendships = append(validFriendships, friendship)
			continue
		}

		// Check if anon profile has a real association
		realProfileID, err := fs.profileService.GetRealProfileID(ctx, friendProfile.ID)
		if err != nil {
			return nil, err
		}

		// Skip if there's a real profile (migration case)
		if realProfileID != uuid.Nil {
			continue
		}

		validFriendships = append(validFriendships, friendship)
	}

	mapperFunc := func(friendship users.Friendship) (dto.FriendshipResponse, error) {
		return mapper.FriendshipToResponse(profile.ID, friendship)
	}

	return ezutil.MapSliceWithError(validFriendships, mapperFunc)
}

func (fs *friendshipServiceImpl) GetDetails(ctx context.Context, profileID, friendshipID uuid.UUID) (dto.FriendDetails, error) {
	profile, err := fs.profileService.GetByID(ctx, profileID)
	if err != nil {
		return dto.FriendDetails{}, err
	}

	spec := users.FriendshipSpecification{}
	spec.Model.ID = friendshipID
	spec.PreloadRelations = []string{"Profile1", "Profile2"}
	friendship, err := fs.friendshipRepository.FindFirstBySpec(ctx, spec)
	if err != nil {
		return dto.FriendDetails{}, err
	}
	if friendship.IsZero() {
		return dto.FriendDetails{}, ungerr.NotFoundError("friendship not found")
	}

	return mapper.MapToFriendDetails(profile.ID, friendship)
}

func (fs *friendshipServiceImpl) IsFriends(ctx context.Context, profileID1, profileID2 uuid.UUID) (bool, bool, error) {
	friendship, err := fs.friendshipRepository.FindByProfileIDs(ctx, profileID1, profileID2)
	if err != nil {
		return false, false, err
	}

	if friendship.IsZero() {
		return false, false, nil
	}

	return true, friendship.Type == users.Anonymous, nil
}

func (fs *friendshipServiceImpl) CreateReal(ctx context.Context, userProfileID, friendProfileID uuid.UUID) (dto.FriendshipResponse, error) {
	var response dto.FriendshipResponse
	err := fs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		profiles, err := fs.profileService.GetByIDs(ctx, []uuid.UUID{userProfileID, friendProfileID})
		if err != nil {
			return err
		}

		userProfile := profiles[userProfileID]
		friendProfile := profiles[friendProfileID]

		newFriendship, err := mapper.OrderProfilesToFriendship(userProfile, friendProfile)
		if err != nil {
			return err
		}

		newFriendship.Type = users.Real

		insertedFriendship, err := fs.friendshipRepository.Insert(ctx, newFriendship)
		if err != nil {
			return err
		}

		response, err = mapper.FriendshipToResponse(userProfile.ID, insertedFriendship)
		return err
	})
	return response, err
}
