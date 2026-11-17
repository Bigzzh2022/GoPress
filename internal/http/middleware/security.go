package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gopress/internal/http/response"
)

// SecurityHeaders adds defensive response headers to every API response so a
// browser never sniffs content types, frames the API, or leaks referrers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		header.Set("Cross-Origin-Resource-Policy", "same-site")
		header.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		c.Next()
	}
}

// BodyLimit rejects oversized payloads before they are read into memory and
// caps the reader so chunked requests cannot bypass the Content-Length check.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			response.Error(c, http.StatusRequestEntityTooLarge, response.CodeBadRequest, "request body too large")
			c.Abort()
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
