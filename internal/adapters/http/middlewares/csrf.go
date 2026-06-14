package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ungerr"
)

const csrfTokenCookieName = "csrf_token"

func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		csrfCookie, err := c.Cookie(csrfTokenCookieName)
		if err != nil || csrfCookie == "" {
			_ = c.Error(ungerr.ForbiddenError("missing CSRF token"))
			c.Abort()
			return
		}

		csrfHeader := c.GetHeader("X-CSRF-Token")
		if csrfHeader == "" || csrfHeader != csrfCookie {
			_ = c.Error(ungerr.ForbiddenError("invalid CSRF token"))
			c.Abort()
			return
		}

		c.Next()
	}
}
