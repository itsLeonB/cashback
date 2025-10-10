package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/domain/profile"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/mapper"
	"github.com/itsLeonB/orcashtrator/internal/util"
)

type profileServiceGrpc struct {
	profileClient profile.ProfileClient
}

func NewProfileService(
	profileClient profile.ProfileClient,
) ProfileService {
	return &profileServiceGrpc{
		profileClient,
	}
}

func (ps *profileServiceGrpc) GetByID(ctx context.Context, id uuid.UUID) (dto.ProfileResponse, error) {
	profile, err := ps.profileClient.Get(ctx, id)
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	return mapper.ProfileToResponse(profile), nil
}

func (ps *profileServiceGrpc) GetNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	profiles, err := ps.profileClient.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	namesByProfileID := make(map[uuid.UUID]string, len(profiles))
	for _, profile := range profiles {
		namesByProfileID[profile.ID] = profile.Name
	}

	return namesByProfileID, nil
}

func (ps *profileServiceGrpc) Update(ctx context.Context, id uuid.UUID, name string) (dto.ProfileResponse, error) {
	request := profile.UpdateRequest{
		ID:   id,
		Name: name,
	}

	updatedProfile, err := ps.profileClient.Update(ctx, request)
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	return mapper.ProfileToResponse(updatedProfile), nil
}

func (ps *profileServiceGrpc) Search(ctx context.Context, profileID uuid.UUID, input string) ([]dto.ProfileResponse, error) {
	if util.IsValidEmail(input) {
		profile, err := ps.profileClient.GetByEmail(ctx, input)
		if err != nil {
			return nil, err
		}
		if profile.ID == profileID {
			return []dto.ProfileResponse{}, nil
		}
		return []dto.ProfileResponse{mapper.ProfileToResponse(profile)}, nil
	}

	profiles, err := ps.profileClient.SearchByName(ctx, input, 10)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProfileResponse, 0, len(profiles))
	for _, profile := range profiles {
		if profile.ID != profileID {
			responses = append(responses, mapper.ProfileToResponse(profile))
		}
	}

	return responses, nil
}
