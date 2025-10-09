package profile

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cocoon-protos/gen/go/profile/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/rotisserie/eris"
)

func FromProfileProto(p *profile.ProfileResponse) (Profile, error) {
	if p == nil {
		return Profile{}, eris.New("profile response is nil")
	}
	profile := p.GetProfile()
	if profile == nil {
		return Profile{}, eris.New("profile is nil")
	}

	userID, err := ezutil.Parse[uuid.UUID](profile.GetUserId())
	if err != nil {
		return Profile{}, err
	}

	amp := p.GetAuditMetadata()
	if amp == nil {
		return Profile{}, eris.New("audit metadata is nil")
	}

	id, err := ezutil.Parse[uuid.UUID](amp.GetId())
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		ID:        id,
		UserID:    userID,
		Name:      profile.GetName(),
		Avatar:    profile.GetAvatar(),
		Email:     profile.GetEmail(),
		CreatedAt: ezutil.FromProtoTime(amp.GetCreatedAt()),
		UpdatedAt: ezutil.FromProtoTime(amp.GetUpdatedAt()),
		DeletedAt: ezutil.FromProtoTime(amp.GetDeletedAt()),
	}, nil
}
