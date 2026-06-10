package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupCSRFRouter() *gin.Engine {
	r := gin.New()
	r.Use(CSRF())
	r.POST("/action", func(c *gin.Context) { c.String(http.StatusOK, "executed") })
	r.GET("/read", func(c *gin.Context) { c.String(http.StatusOK, "executed") })
	return r
}

func TestCSRF_GETRequest_Skipped(t *testing.T) {
	r := setupCSRFRouter()
	req := httptest.NewRequest(http.MethodGet, "/read", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "executed", w.Body.String())
}

func TestCSRF_POSTRequest_MissingCookie(t *testing.T) {
	r := setupCSRFRouter()
	req := httptest.NewRequest(http.MethodPost, "/action", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.NotContains(t, w.Body.String(), "executed")
}

func TestCSRF_POSTRequest_MissingHeader(t *testing.T) {
	r := setupCSRFRouter()
	req := httptest.NewRequest(http.MethodPost, "/action", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token123"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.NotContains(t, w.Body.String(), "executed")
}

func TestCSRF_POSTRequest_Mismatch(t *testing.T) {
	r := setupCSRFRouter()
	req := httptest.NewRequest(http.MethodPost, "/action", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token123"})
	req.Header.Set("X-CSRF-Token", "wrong")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.NotContains(t, w.Body.String(), "executed")
}

func TestCSRF_POSTRequest_Valid(t *testing.T) {
	r := setupCSRFRouter()
	req := httptest.NewRequest(http.MethodPost, "/action", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token123"})
	req.Header.Set("X-CSRF-Token", "token123")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "executed", w.Body.String())
}
