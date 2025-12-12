package friendship

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cocoon-protos/gen/go/friendship/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/rotisserie/eris"
)

func fromFriendshipTypeProto(ft friendship.FriendshipType) (appconstant.FriendshipType, error) {
	switch ft {
	case friendship.FriendshipType_FRIENDSHIP_TYPE_ANON:
		return appconstant.Anonymous, nil
	case friendship.FriendshipType_FRIENDSHIP_TYPE_REAL:
		return appconstant.Real, nil
	default:
		return "", eris.Errorf("unknown friendship type constant: %s", ft.String())
	}
}

func fromFriendshipProto(f *friendship.FriendshipResponse) (Friendship, error) {
	if f == nil {
		return Friendship{}, eris.New("friendship is nil")
	}

	id, err := ezutil.Parse[uuid.UUID](f.GetId())
	if err != nil {
		return Friendship{}, err
	}

	friendshipType, err := fromFriendshipTypeProto(f.GetType())
	if err != nil {
		return Friendship{}, err
	}

	profileID, err := ezutil.Parse[uuid.UUID](f.GetProfileId())
	if err != nil {
		return Friendship{}, err
	}

	return Friendship{
		ID:            id,
		Type:          friendshipType,
		ProfileID:     profileID,
		ProfileName:   f.GetProfileName(),
		ProfileAvatar: f.GetProfileAvatar(),
		CreatedAt:     ezutil.FromProtoTime(f.GetCreatedAt()),
		UpdatedAt:     ezutil.FromProtoTime(f.GetUpdatedAt()),
		DeletedAt:     ezutil.FromProtoTime(f.GetDeletedAt()),
	}, nil
}
