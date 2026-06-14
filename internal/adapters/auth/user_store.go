package authadapter

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/go-crud"
	"golang.org/x/text/currency"
)

type userStoreAdapter struct {
	userRepo   repository.UserRepository
	profileSvc service.ProfileService
}

func NewUserStore(userRepo repository.UserRepository, profileSvc service.ProfileService) auth.UserStore {
	return &userStoreAdapter{userRepo, profileSvc}
}

func (a *userStoreAdapter) FindByID(ctx context.Context, userID string) (auth.User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return auth.User{}, err
	}
	spec := crud.Specification[users.User]{}
	spec.Model.ID = uid
	user, err := a.userRepo.FindFirst(ctx, spec)
	if err != nil {
		return auth.User{}, err
	}
	if user.IsZero() {
		return auth.User{}, auth.ErrUserNotFound
	}
	return toAuthUser(user), nil
}

func (a *userStoreAdapter) FindByEmail(ctx context.Context, email string) (auth.User, error) {
	spec := crud.Specification[users.User]{}
	spec.Model.Email = email
	user, err := a.userRepo.FindFirst(ctx, spec)
	if err != nil {
		return auth.User{}, err
	}
	if user.IsZero() {
		return auth.User{}, auth.ErrUserNotFound
	}
	return toAuthUser(user), nil
}

func (a *userStoreAdapter) Create(ctx context.Context, email, passwordHash string) (auth.User, error) {
	user, err := a.userRepo.Insert(ctx, users.User{
		Email:    email,
		Password: passwordHash,
	})
	if err != nil {
		return auth.User{}, err
	}
	return toAuthUser(user), nil
}

func (a *userStoreAdapter) CreateOAuth(ctx context.Context, email, name, avatar string) (auth.User, error) {
	user, err := a.userRepo.Insert(ctx, users.User{
		Email: email,
		VerifiedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	})
	if err != nil {
		return auth.User{}, err
	}

	if _, err = a.profileSvc.Create(ctx, dto.NewProfileRequest{
		UserID:       user.ID,
		Name:         name,
		Avatar:       avatar,
		HomeCurrency: currency.IDR.String(),
	}); err != nil {
		return auth.User{}, err
	}

	return toAuthUser(user), nil
}

func (a *userStoreAdapter) SetVerified(ctx context.Context, userID string, name, avatar string) (auth.User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return auth.User{}, err
	}

	spec := crud.Specification[users.User]{}
	spec.Model.ID = uid
	spec.PreloadRelations = []string{"Profile"}
	user, err := a.userRepo.FindFirst(ctx, spec)
	if err != nil {
		return auth.User{}, err
	}
	if user.IsZero() {
		return auth.User{}, auth.ErrUserNotFound
	}

	if user.Profile.IsZero() {
		if _, err = a.profileSvc.Create(ctx, dto.NewProfileRequest{
			UserID:       user.ID,
			Name:         name,
			Avatar:       avatar,
			HomeCurrency: currency.IDR.String(),
		}); err != nil {
			return auth.User{}, err
		}
		// Refresh profile after creation
		updated, err := a.userRepo.FindFirst(ctx, spec)
		if err != nil {
			return auth.User{}, err
		}
		user = updated
	}

	user.VerifiedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	updated, err := a.userRepo.Update(ctx, user)
	if err != nil {
		return auth.User{}, err
	}
	return toAuthUser(updated), nil
}

func (a *userStoreAdapter) Exists(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	spec := crud.Specification[users.User]{}
	spec.Model.ID = uid
	user, err := a.userRepo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	if user.IsZero() {
		return auth.ErrUserNotFound
	}
	return nil
}

func (a *userStoreAdapter) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	spec := crud.Specification[users.User]{}
	spec.Model.ID = uid
	u, err := a.userRepo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	if u.IsZero() {
		return auth.ErrUserNotFound
	}

	u.Password = passwordHash
	_, err = a.userRepo.Update(ctx, u)
	return err
}

func toAuthUser(u users.User) auth.User {
	profileID := ""
	if !u.Profile.IsZero() {
		profileID = u.Profile.ID.String()
	}
	return auth.User{
		ID:           u.ID.String(),
		Email:        u.Email,
		PasswordHash: u.Password,
		Verified:     u.IsVerified(),
		ProfileID:    profileID,
	}
}
