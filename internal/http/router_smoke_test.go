package httpapi

import (
	"strings"
	"testing"

	"gorm.io/gorm"

	"gopress/internal/config"
	"gopress/internal/platform/cache"
)

func TestRouteRegistrationSmoke(t *testing.T) {
	cfg := config.Config{
		JWTSecret:   strings.Repeat("a", config.MinJWTSecretLength),
		CORSOrigins: []string{"http://localhost:3000"},
	}
	router := NewRouter(cfg, &gorm.DB{}, &cache.Cache{})
	if router == nil {
		t.Fatal("router was not created")
	}
}
