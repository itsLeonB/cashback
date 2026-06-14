package authadapter

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/service/cache"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
)

type sessionCacheAdapter struct {
	inner cache.Cache[uuid.UUID]
}

func NewSessionCacheAdapter(inner cache.Cache[uuid.UUID]) auth.SessionCache {
	return &sessionCacheAdapter{inner}
}

func (a *sessionCacheAdapter) Get(sessionID string, loader func(string) (string, bool)) (string, bool) {
	// The inner cache stores uuid.Nil on loader failure; we only use the value
	// when hit==true, so the zero UUID never propagates to callers.
	userID, hit := a.inner.Get(sessionID, func(sid string) (uuid.UUID, bool) {
		userIDStr, ok := loader(sid)
		if !ok {
			return uuid.Nil, false
		}
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, false
		}
		return uid, true
	})
	if !hit {
		return "", false
	}
	return userID.String(), true
}

func (a *sessionCacheAdapter) Delete(sessionID string) {
	a.inner.Delete(sessionID)
}

func (a *sessionCacheAdapter) Shutdown() error {
	return a.inner.Shutdown()
}
