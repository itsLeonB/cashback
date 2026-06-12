package service_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/service/cache"
	authadapter "github.com/itsLeonB/cashback/internal/adapters/auth"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/cashback/internal/mocks"
	"github.com/itsLeonB/sekure"
	"github.com/itsLeonB/ungerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestAuthService(jwtSvc sekure.JWTService, sessionSvc service.SessionService, sessionCache cache.Cache[uuid.UUID]) service.AuthService {
	jwtAdapter := authadapter.NewJWTService(jwtSvc)
	cacheAdapter := authadapter.NewSessionCacheAdapter(sessionCache)
	return service.NewAuthService(jwtAdapter, nil, nil, nil, nil, "", "", nil, sessionSvc, cacheAdapter, service.AuthHooks{})
}

func TestVerifyToken_Success(t *testing.T) {
	jwtMock := mocks.NewMockJWTService(t)
	sessionMock := mocks.NewMockSessionService(t)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](time.Hour)

	userID := uuid.New()
	sessionID := uuid.New()
	profileID := uuid.New()

	rawFgp := "test-fingerprint"
	h := sha256.Sum256([]byte(rawFgp))
	fgpHash := hex.EncodeToString(h[:])

	claims := sekure.JWTClaims{
		Data: map[string]any{
			appconstant.ContextUserID.String():      userID.String(),
			appconstant.ContextSessionID.String():   sessionID.String(),
			appconstant.ContextFingerprint.String(): fgpHash,
			appconstant.ContextProfileID.String():   profileID.String(),
		},
	}

	jwtMock.EXPECT().VerifyToken("valid-token").Return(claims, nil)
	sessionMock.EXPECT().GetByID(mock.Anything, sessionID.String()).Return(auth.Session{
		ID:     sessionID.String(),
		UserID: userID.String(),
	}, nil)

	svc := newTestAuthService(jwtMock, sessionMock, sessionCache)

	valid, data, err := svc.VerifyToken(context.Background(), "valid-token", rawFgp)

	assert.NoError(t, err)
	assert.True(t, valid)
	assert.Equal(t, profileID, data[appconstant.ContextProfileID.String()])
	assert.Equal(t, sessionID, data[appconstant.ContextSessionID.String()])
}

func TestVerifyToken_InvalidFingerprint(t *testing.T) {
	jwtMock := mocks.NewMockJWTService(t)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](time.Hour)

	userID := uuid.New()
	sessionID := uuid.New()

	claims := sekure.JWTClaims{
		Data: map[string]any{
			appconstant.ContextUserID.String():      userID.String(),
			appconstant.ContextSessionID.String():   sessionID.String(),
			appconstant.ContextFingerprint.String(): "expected-hash-in-token",
		},
	}

	jwtMock.EXPECT().VerifyToken("token").Return(claims, nil)

	svc := newTestAuthService(jwtMock, nil, sessionCache)

	valid, data, err := svc.VerifyToken(context.Background(), "token", "wrong-fingerprint")

	assert.Error(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
	var appErr ungerr.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, "invalid token fingerprint", appErr.Details())
}

func TestVerifyToken_MissingFingerprintClaim(t *testing.T) {
	jwtMock := mocks.NewMockJWTService(t)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](time.Hour)

	claims := sekure.JWTClaims{
		Data: map[string]any{
			appconstant.ContextUserID.String():    uuid.New().String(),
			appconstant.ContextSessionID.String(): uuid.New().String(),
		},
	}

	jwtMock.EXPECT().VerifyToken("token").Return(claims, nil)

	svc := newTestAuthService(jwtMock, nil, sessionCache)

	valid, data, err := svc.VerifyToken(context.Background(), "token", "any-fingerprint")

	assert.Error(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
	var appErr ungerr.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, "missing fingerprint claim", appErr.Details())
}

func TestVerifyToken_InvalidToken(t *testing.T) {
	jwtMock := mocks.NewMockJWTService(t)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](time.Hour)

	jwtMock.EXPECT().VerifyToken("bad-token").Return(sekure.JWTClaims{}, errors.New("invalid token"))

	svc := newTestAuthService(jwtMock, nil, sessionCache)

	valid, data, err := svc.VerifyToken(context.Background(), "bad-token", "")

	assert.Error(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
}

func TestVerifyToken_SessionNotFound(t *testing.T) {
	jwtMock := mocks.NewMockJWTService(t)
	sessionMock := mocks.NewMockSessionService(t)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](time.Hour)

	userID := uuid.New()
	sessionID := uuid.New()
	rawFgp := "fgp"
	h := sha256.Sum256([]byte(rawFgp))
	fgpHash := hex.EncodeToString(h[:])

	claims := sekure.JWTClaims{
		Data: map[string]any{
			appconstant.ContextUserID.String():      userID.String(),
			appconstant.ContextSessionID.String():   sessionID.String(),
			appconstant.ContextFingerprint.String(): fgpHash,
		},
	}

	jwtMock.EXPECT().VerifyToken("token").Return(claims, nil)
	sessionMock.EXPECT().GetByID(mock.Anything, sessionID.String()).Return(auth.Session{}, errors.New("not found"))

	svc := newTestAuthService(jwtMock, sessionMock, sessionCache)

	valid, data, err := svc.VerifyToken(context.Background(), "token", rawFgp)

	assert.Error(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
	var appErr ungerr.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, "session is not found", appErr.Details())
}

func TestVerifyToken_SessionBelongsToDifferentUser(t *testing.T) {
	jwtMock := mocks.NewMockJWTService(t)
	sessionMock := mocks.NewMockSessionService(t)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](time.Hour)

	attackerUserID := uuid.New()
	victimUserID := uuid.New()
	sessionID := uuid.New()
	rawFgp := "fgp"
	h := sha256.Sum256([]byte(rawFgp))
	fgpHash := hex.EncodeToString(h[:])

	claims := sekure.JWTClaims{
		Data: map[string]any{
			appconstant.ContextUserID.String():      victimUserID.String(),
			appconstant.ContextSessionID.String():   sessionID.String(),
			appconstant.ContextFingerprint.String(): fgpHash,
		},
	}

	jwtMock.EXPECT().VerifyToken("forged-token").Return(claims, nil)
	sessionMock.EXPECT().GetByID(mock.Anything, sessionID.String()).Return(auth.Session{
		ID:     sessionID.String(),
		UserID: attackerUserID.String(),
	}, nil)

	svc := newTestAuthService(jwtMock, sessionMock, sessionCache)

	valid, data, err := svc.VerifyToken(context.Background(), "forged-token", rawFgp)

	assert.Error(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
	var appErr ungerr.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, "session does not belong to user", appErr.Details())
}

func TestVerifyToken_CachedSession(t *testing.T) {
	jwtMock := mocks.NewMockJWTService(t)
	sessionMock := mocks.NewMockSessionService(t)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](time.Hour)

	userID := uuid.New()
	sessionID := uuid.New()
	profileID := uuid.New()
	rawFgp := "fgp"
	h := sha256.Sum256([]byte(rawFgp))
	fgpHash := hex.EncodeToString(h[:])

	claims := sekure.JWTClaims{
		Data: map[string]any{
			appconstant.ContextUserID.String():      userID.String(),
			appconstant.ContextSessionID.String():   sessionID.String(),
			appconstant.ContextFingerprint.String(): fgpHash,
			appconstant.ContextProfileID.String():   profileID.String(),
		},
	}

	jwtMock.EXPECT().VerifyToken("token").Return(claims, nil)
	sessionMock.EXPECT().GetByID(mock.Anything, sessionID.String()).Return(auth.Session{
		ID:     sessionID.String(),
		UserID: userID.String(),
	}, nil).Once()

	svc := newTestAuthService(jwtMock, sessionMock, sessionCache)

	// First call - hits DB via fallback
	valid, _, err := svc.VerifyToken(context.Background(), "token", rawFgp)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Second call - uses cache, no additional GetByID on session
	valid, _, err = svc.VerifyToken(context.Background(), "token", rawFgp)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestVerifyToken_MissingUserID(t *testing.T) {
	jwtMock := mocks.NewMockJWTService(t)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](time.Hour)

	rawFgp := "fgp"
	h := sha256.Sum256([]byte(rawFgp))
	fgpHash := hex.EncodeToString(h[:])

	claims := sekure.JWTClaims{
		Data: map[string]any{
			appconstant.ContextSessionID.String():   uuid.New().String(),
			appconstant.ContextFingerprint.String(): fgpHash,
		},
	}

	jwtMock.EXPECT().VerifyToken("token").Return(claims, nil)

	svc := newTestAuthService(jwtMock, nil, sessionCache)

	valid, data, err := svc.VerifyToken(context.Background(), "token", rawFgp)

	assert.Error(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
}

func TestVerifyToken_MissingSessionID(t *testing.T) {
	jwtMock := mocks.NewMockJWTService(t)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](time.Hour)

	userID := uuid.New()
	rawFgp := "fgp"
	h := sha256.Sum256([]byte(rawFgp))
	fgpHash := hex.EncodeToString(h[:])

	claims := sekure.JWTClaims{
		Data: map[string]any{
			appconstant.ContextUserID.String():      userID.String(),
			appconstant.ContextFingerprint.String(): fgpHash,
		},
	}

	jwtMock.EXPECT().VerifyToken("token").Return(claims, nil)

	svc := newTestAuthService(jwtMock, nil, sessionCache)

	valid, data, err := svc.VerifyToken(context.Background(), "token", rawFgp)

	assert.Error(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
}
