package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newTestRouter(middleware ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware...)
	router.POST("/write", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	router.POST("/consume", func(c *gin.Context) {
		// Mirrors what ShouldBindJSON does: the cap only bites when the body is read.
		if _, err := io.Copy(io.Discard, c.Request.Body); err != nil {
			c.String(http.StatusRequestEntityTooLarge, "too large")
			return
		}
		c.String(http.StatusOK, "ok")
	})
	router.GET("/read", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return router
}

func TestRateLimiterThrottlesWithoutRedis(t *testing.T) {
	limiter := NewRateLimiter(nil, "test", 2, time.Minute)
	router := newTestRouter(Limit(limiter))

	for attempt := 1; attempt <= 2; attempt++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/read", nil)
		request.RemoteAddr = "203.0.113.7:1234"
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("attempt %d: status = %d, want 200", attempt, recorder.Code)
		}
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/read", nil)
	request.RemoteAddr = "203.0.113.7:1234"
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("throttled status = %d, want 429", recorder.Code)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("429 response is missing Retry-After")
	}

	limiter.Reset(request.Context(), "203.0.113.7")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/read", nil))
	if recorder.Code == http.StatusTooManyRequests {
		t.Fatal("Reset() did not clear the counter")
	}
}

func TestCORSDoesNotCombineWildcardWithCredentials(t *testing.T) {
	router := newTestRouter(CORS([]string{"*"}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/read", nil)
	request.Header.Set("Origin", "https://evil.example")
	router.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want empty for wildcard origins", got)
	}
}

func TestCORSAllowlistOnlyEchoesKnownOrigins(t *testing.T) {
	router := newTestRouter(CORS([]string{"http://localhost:5173"}))

	allowed := httptest.NewRecorder()
	allowedRequest := httptest.NewRequest(http.MethodGet, "/read", nil)
	allowedRequest.Header.Set("Origin", "http://localhost:5173")
	router.ServeHTTP(allowed, allowedRequest)
	if got := allowed.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allowed origin header = %q, want the configured origin", got)
	}
	if got := allowed.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("credentials header = %q, want true", got)
	}

	blocked := httptest.NewRecorder()
	blockedRequest := httptest.NewRequest(http.MethodGet, "/read", nil)
	blockedRequest.Header.Set("Origin", "https://evil.example")
	router.ServeHTTP(blocked, blockedRequest)
	if got := blocked.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("disallowed origin was echoed back: %q", got)
	}
}

func TestRequireTrustedOrigin(t *testing.T) {
	router := newTestRouter(RequireTrustedOrigin([]string{"http://localhost:5173"}))

	cases := []struct {
		name   string
		origin string
		status int
	}{
		{"missing origin", "", http.StatusOK},
		{"allowed origin", "http://localhost:5173", http.StatusOK},
		{"cross-site origin", "https://evil.example", http.StatusForbidden},
		{"null origin", "null", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/write", nil)
			if tc.origin != "" {
				request.Header.Set("Origin", tc.origin)
			}
			router.ServeHTTP(recorder, request)
			if recorder.Code != tc.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.status)
			}
		})
	}
}

func TestBodyLimitRejectsOversizedPayload(t *testing.T) {
	router := newTestRouter(BodyLimit(16))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/write", strings.NewReader(strings.Repeat("a", 64)))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", recorder.Code)
	}

	chunked := httptest.NewRecorder()
	chunkedRequest := httptest.NewRequest(http.MethodPost, "/consume", strings.NewReader(strings.Repeat("a", 64)))
	chunkedRequest.ContentLength = -1 // forces MaxBytesReader to enforce the cap
	router.ServeHTTP(chunked, chunkedRequest)
	if chunked.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("chunked status = %d, want 413", chunked.Code)
	}
}
