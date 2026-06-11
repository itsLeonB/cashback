package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAuthRouter(authMock *mocks.MockAuthService) *gin.Engine {
	r := gin.New()
	r.Use(newCookieAuthMiddleware(authMock))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"profileID": c.GetString("profileID")})
	})
	return r
}

func TestCookieAuthMiddleware_Success(t *testing.T) {
	authMock := mocks.NewMockAuthService(t)
	authMock.EXPECT().VerifyToken(mock.Anything, "valid-token", "").Return(true, map[string]any{"profileID": "abc"}, nil)

	r := setupAuthRouter(authMock)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "valid-token"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "abc")
}

func TestCookieAuthMiddleware_NoCookie(t *testing.T) {
	authMock := mocks.NewMockAuthService(t)

	r := setupAuthRouter(authMock)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Handler should not execute - no profileID in response
	assert.NotContains(t, w.Body.String(), "profileID")
}

func TestCookieAuthMiddleware_InvalidToken(t *testing.T) {
	authMock := mocks.NewMockAuthService(t)
	authMock.EXPECT().VerifyToken(mock.Anything, "bad-token", "").Return(false, nil, errors.New("invalid"))

	r := setupAuthRouter(authMock)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "bad-token"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.NotContains(t, w.Body.String(), "profileID")
}

func TestCookieAuthMiddleware_NoCookieNoHeader_Rejected(t *testing.T) {
	authMock := mocks.NewMockAuthService(t)

	r := setupAuthRouter(authMock)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// No Authorization header, no cookie
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.NotContains(t, w.Body.String(), "profileID")
}
