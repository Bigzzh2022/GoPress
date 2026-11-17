package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"gopress/internal/http/response"
)

// RequireTrustedOrigin blocks state-changing requests that a browser marks as
// coming from another site. The refresh cookie is the only ambient credential,
// so cross-site refresh/logout are the interesting CSRF targets.
func RequireTrustedOrigin(origins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	allowAny := false
	for _, origin := range origins {
		origin = normalizeOrigin(origin)
		if origin == "*" {
			allowAny = true
			continue
		}
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		if allowAny {
			c.Next()
			return
		}
		origin := normalizeOrigin(c.GetHeader("Origin"))
		if origin == "" {
			origin = originFromReferer(c.GetHeader("Referer"))
		}
		// Non-browser clients such as curl send neither header; they also do not
		// carry ambient cookies, so they are not a CSRF vector.
		if origin == "" {
			c.Next()
			return
		}
		if _, ok := allowed[origin]; !ok {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "untrusted request origin")
			c.Abort()
			return
		}
		c.Next()
	}
}

func normalizeOrigin(value string) string {
	return strings.ToLower(strings.TrimRight(strings.TrimSpace(value), "/"))
}

func originFromReferer(referer string) string {
	referer = strings.TrimSpace(referer)
	if referer == "" {
		return ""
	}
	parsed, err := url.Parse(referer)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s", strings.ToLower(parsed.Scheme), strings.ToLower(parsed.Host))
}
