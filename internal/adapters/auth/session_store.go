package authadapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-authkit"
	"github.com/itsLeonB/go-crud"
)

type sessionStoreAdapter struct {
	repo crud.Repository[users.Session]
}

func NewSessionStore(repo crud.Repository[users.Session]) authkit.SessionStore {
	return &sessionStoreAdapter{repo}
}

func (a *sessionStoreAdapter) Create(ctx context.Context, userID string) (authkit.Session, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return authkit.Session{}, err
	}

	session, err := a.repo.Insert(ctx, users.Session{
		UserID:     uid,
		LastUsedAt: time.Now(),
	})
	if err != nil {
		return authkit.Session{}, err
	}
	return toAuthSession(session), nil
}

func (a *sessionStoreAdapter) GetByID(ctx context.Context, id string) (authkit.Session, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return authkit.Session{}, err
	}

	spec := crud.Specification[users.Session]{}
	spec.Model.ID = uid
	session, err := a.repo.FindFirst(ctx, spec)
	if err != nil {
		return authkit.Session{}, err
	}
	if session.IsZero() {
		return authkit.Session{}, authkit.ErrSessionNotFound
	}
	return toAuthSession(session), nil
}

func (a *sessionStoreAdapter) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	spec := crud.Specification[users.Session]{}
	spec.Model.ID = uid
	session, err := a.repo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	if session.IsZero() {
		return nil
	}
	return a.repo.Delete(ctx, session)
}

func (a *sessionStoreAdapter) Touch(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	spec := crud.Specification[users.Session]{}
	spec.Model.ID = uid
	session, err := a.repo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	if session.IsZero() {
		return nil
	}

	session.LastUsedAt = time.Now()
	_, err = a.repo.Update(ctx, session)
	return err
}

func toAuthSession(s users.Session) authkit.Session {
	return authkit.Session{
		ID:     s.ID.String(),
		UserID: s.UserID.String(),
	}
}
