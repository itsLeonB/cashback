package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/admin"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/sekure"
	"github.com/itsLeonB/ungerr"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) error
	Login(ctx context.Context, req dto.InternalLoginRequest) (dto.TokenResponse, error)
	VerifyToken(ctx context.Context, token string) (bool, map[string]any, error)
}

type authService struct {
	userRepo    crud.Repository[admin.User]
	hashService sekure.HashService
	jwtService  sekure.JWTService
}

func NewAuthService(
	userRepo crud.Repository[admin.User],
	hashService sekure.HashService,
	jwtService sekure.JWTService,
) *authService {
	return &authService{
		userRepo,
		hashService,
		jwtService,
	}
}

func (as *authService) Register(ctx context.Context, req dto.RegisterRequest) error {
	users, err := as.userRepo.FindAll(ctx, crud.Specification[admin.User]{})
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return ungerr.ForbiddenError("cannot register as there exists admin users")
	}

	hash, err := as.hashService.Hash(req.Password)
	if err != nil {
		return err
	}

	newUser := admin.User{
		Email:    req.Email,
		Password: hash,
	}

	_, err = as.userRepo.Insert(ctx, newUser)
	return err
}

func (as *authService) Login(ctx context.Context, req dto.InternalLoginRequest) (dto.TokenResponse, error) {
	spec := crud.Specification[admin.User]{}
	spec.Model.Email = req.Email
	user, err := as.userRepo.FindFirst(ctx, spec)
	if err != nil {
		return dto.TokenResponse{}, err
	}
	if user.IsZero() {
		return dto.TokenResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}

	ok, err := as.hashService.CheckHash(user.Password, req.Password)
	if err != nil {
		return dto.TokenResponse{}, err
	}
	if !ok {
		return dto.TokenResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}

	token, err := as.jwtService.CreateToken(map[string]any{
		appconstant.ContextUserID.String(): user.ID,
	})
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.NewTokenResp(token, ""), nil
}

func (as *authService) VerifyToken(ctx context.Context, token string) (bool, map[string]any, error) {
	claims, err := as.jwtService.VerifyToken(token)
	if err != nil {
		return false, nil, err
	}

	tokenUserId, exists := claims.Data[appconstant.ContextUserID.String()]
	if !exists {
		return false, nil, ungerr.Unknown("missing user ID from token")
	}
	stringUserID, ok := tokenUserId.(string)
	if !ok {
		return false, nil, ungerr.Unknown("error asserting userID, is not a string")
	}
	userID, err := ezutil.Parse[uuid.UUID](stringUserID)
	if err != nil {
		return false, nil, err
	}

	spec := crud.Specification[admin.User]{}
	spec.Model.ID = userID
	user, err := as.userRepo.FindFirst(ctx, spec)
	if err != nil {
		return false, nil, err
	}
	if user.IsZero() {
		return false, nil, ungerr.UnauthorizedError("user not found")
	}

	return true, map[string]any{
		appconstant.ContextUserID.String(): user.ID,
	}, nil
}
