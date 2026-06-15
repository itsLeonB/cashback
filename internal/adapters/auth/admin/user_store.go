package adminauth

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/admin"
	"github.com/itsLeonB/go-authkit"
	"github.com/itsLeonB/go-crud"
)

type UserStore struct {
	repo crud.Repository[admin.User]
}

func NewUserStore(repo crud.Repository[admin.User]) *UserStore {
	return &UserStore{repo: repo}
}

func (s *UserStore) FindByID(ctx context.Context, userID string) (authkit.User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return authkit.User{}, authkit.ErrUserNotFound
	}
	spec := crud.Specification[admin.User]{}
	spec.Model.ID = uid
	user, err := s.repo.FindFirst(ctx, spec)
	if err != nil {
		return authkit.User{}, err
	}
	if user.IsZero() {
		return authkit.User{}, authkit.ErrUserNotFound
	}
	return toAuthUser(user), nil
}

func (s *UserStore) FindByEmail(ctx context.Context, email string) (authkit.User, error) {
	spec := crud.Specification[admin.User]{}
	spec.Model.Email = email
	user, err := s.repo.FindFirst(ctx, spec)
	if err != nil {
		return authkit.User{}, err
	}
	if user.IsZero() {
		return authkit.User{}, authkit.ErrUserNotFound
	}
	return toAuthUser(user), nil
}

func (s *UserStore) Create(ctx context.Context, email, passwordHash string) (authkit.User, error) {
	user, err := s.repo.Insert(ctx, admin.User{
		Email:    email,
		Password: passwordHash,
	})
	if err != nil {
		return authkit.User{}, err
	}
	return toAuthUser(user), nil
}

func (s *UserStore) CreateOAuth(_ context.Context, _, _, _ string) (authkit.User, error) {
	return authkit.User{}, authkit.ErrNotSupported
}

func (s *UserStore) SetVerified(_ context.Context, userID, _, _ string) (authkit.User, error) {
	// Admin users are always considered verified.
	return authkit.User{ID: userID, Verified: true}, nil
}

func (s *UserStore) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return authkit.ErrUserNotFound
	}
	spec := crud.Specification[admin.User]{}
	spec.Model.ID = uid
	user, err := s.repo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	if user.IsZero() {
		return authkit.ErrUserNotFound
	}
	user.Password = passwordHash
	_, err = s.repo.Update(ctx, user)
	return err
}

func (s *UserStore) Exists(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return authkit.ErrUserNotFound
	}
	spec := crud.Specification[admin.User]{}
	spec.Model.ID = uid
	user, err := s.repo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	if user.IsZero() {
		return authkit.ErrUserNotFound
	}
	return nil
}

func toAuthUser(u admin.User) authkit.User {
	return authkit.User{
		ID:           u.ID.String(),
		Email:        u.Email,
		PasswordHash: u.Password,
		Verified:     true, // admin users are always verified
	}
}
