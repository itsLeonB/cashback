package cookie

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessTokenName  = "access_token"
	RefreshTokenName = "refresh_token"
	CSRFTokenName    = "csrf_token"
	FingerprintName  = "__Secure-Fgp"
)

type Config struct {
	Domain     string
	Secure     bool
	SameSite   http.SameSite
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func SetAccessToken(c *gin.Context, cfg Config, token string) {
	http.SetCookie(c.Writer, &http.Cookie{ // #nosec G124 -- Secure is set from cfg, true in production
		Name:     AccessTokenName,
		Value:    token,
		Path:     "/api",
		Domain:   cfg.Domain,
		MaxAge:   int(cfg.AccessTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}

func SetRefreshToken(c *gin.Context, cfg Config, token string) {
	http.SetCookie(c.Writer, &http.Cookie{ // #nosec G124 -- Secure is set from cfg, true in production
		Name:     RefreshTokenName,
		Value:    token,
		Path:     "/api/v1/auth",
		Domain:   cfg.Domain,
		MaxAge:   int(cfg.RefreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}

func SetCSRFToken(c *gin.Context, cfg Config, token string) {
	http.SetCookie(c.Writer, &http.Cookie{ // #nosec G124 -- HttpOnly=false intentional: JS reads CSRF token for double-submit
		Name:     CSRFTokenName,
		Value:    token,
		Path:     "/api",
		Domain:   cfg.Domain,
		MaxAge:   int(cfg.AccessTTL.Seconds()),
		HttpOnly: false,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}

func SetFingerprint(c *gin.Context, cfg Config, value string) {
	http.SetCookie(c.Writer, &http.Cookie{ // #nosec G124 -- Secure is set from cfg, true in production
		Name:     FingerprintName,
		Value:    value,
		Path:     "/api",
		Domain:   cfg.Domain,
		MaxAge:   int(cfg.RefreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}

func ClearTokens(c *gin.Context, cfg Config) {
	http.SetCookie(c.Writer, &http.Cookie{ // #nosec G124 -- clearing cookie, Secure from cfg
		Name:     AccessTokenName,
		Value:    "",
		Path:     "/api",
		Domain:   cfg.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{ // #nosec G124 -- clearing cookie, Secure from cfg
		Name:     RefreshTokenName,
		Value:    "",
		Path:     "/api/v1/auth",
		Domain:   cfg.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{ // #nosec G124 -- clearing cookie, HttpOnly=false matches CSRF set cookie
		Name:     CSRFTokenName,
		Value:    "",
		Path:     "/api",
		Domain:   cfg.Domain,
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{ // #nosec G124 -- clearing cookie, Secure from cfg
		Name:     FingerprintName,
		Value:    "",
		Path:     "/api",
		Domain:   cfg.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	})
}
