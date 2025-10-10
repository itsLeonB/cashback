package profile

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/itsLeonB/cocoon-protos/gen/go/profile/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/util"
	"github.com/rotisserie/eris"
	"google.golang.org/grpc"
)

type ProfileClient interface {
	Get(ctx context.Context, id uuid.UUID) (Profile, error)
	Create(ctx context.Context, req CreateRequest) (Profile, error)
	GetByIDs(ctx context.Context, ids uuid.UUIDs) ([]Profile, error)
	Update(ctx context.Context, req UpdateRequest) (Profile, error)
	GetByEmail(ctx context.Context, email string) (Profile, error)
	SearchByName(ctx context.Context, query string, limit int) ([]Profile, error)
}

type profileClient struct {
	validate *validator.Validate
	client   profile.ProfileServiceClient
}

func NewProfileClient(
	validate *validator.Validate,
	conn *grpc.ClientConn,
) ProfileClient {
	if validate == nil {
		panic("validator is nil")
	}

	return &profileClient{
		validate,
		profile.NewProfileServiceClient(conn),
	}
}

func (pc *profileClient) Get(ctx context.Context, id uuid.UUID) (Profile, error) {
	if id == uuid.Nil {
		return Profile{}, eris.New("id is nil")
	}

	request := &profile.GetRequest{
		ProfileId: id.String(),
	}

	response, err := pc.client.Get(ctx, request)
	if err != nil {
		return Profile{}, err
	}

	return FromProfileProto(response.GetProfile())
}

func (pc *profileClient) Create(ctx context.Context, req CreateRequest) (Profile, error) {
	if err := pc.validate.Struct(req); err != nil {
		return Profile{}, err
	}

	request := &profile.CreateRequest{
		Name: req.Name,
	}
	if req.UserID != uuid.Nil {
		userID := req.UserID.String()
		request.UserId = &userID
	}

	response, err := pc.client.Create(ctx, request)
	if err != nil {
		return Profile{}, err
	}

	return FromProfileProto(response.GetProfile())
}

func (pc *profileClient) GetByIDs(ctx context.Context, ids uuid.UUIDs) ([]Profile, error) {
	if ids == nil {
		return nil, eris.New("ids is nil")
	}

	request := &profile.GetByIDsRequest{
		ProfileIds: ezutil.MapSlice(ids, util.ToString),
	}

	response, err := pc.client.GetByIDs(ctx, request)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSliceWithError(response.GetProfiles(), FromProfileProto)
}

func (pc *profileClient) Update(ctx context.Context, req UpdateRequest) (Profile, error) {
	request := &profile.UpdateRequest{
		Id: req.ID.String(),
		Profile: &profile.Profile{
			UserId: req.UserID.String(),
			Name:   req.Name,
			Avatar: req.Avatar,
		},
	}

	response, err := pc.client.Update(ctx, request)
	if err != nil {
		return Profile{}, err
	}

	return FromProfileProto(response.GetProfile())
}

func (pc *profileClient) GetByEmail(ctx context.Context, email string) (Profile, error) {
	request := profile.GetByEmailRequest{
		Email: email,
	}
	response, err := pc.client.GetByEmail(ctx, &request)
	if err != nil {
		return Profile{}, err
	}
	return FromProfileProto(response.GetProfile())
}

func (pc *profileClient) SearchByName(ctx context.Context, query string, limit int) ([]Profile, error) {
	request := profile.SearchByNameRequest{
		Query: query,
		Limit: int64(limit),
	}
	response, err := pc.client.SearchByName(ctx, &request)
	if err != nil {
		return nil, err
	}
	return ezutil.MapSliceWithError(response.GetProfiles(), FromProfileProto)
}
