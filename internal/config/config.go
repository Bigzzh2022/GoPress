package config

import (
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Config struct {
	Port            string
	MySQLDSN        string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CORSOrigins     []string
	TrustedProxies  []string
	CookieSecure    bool
	AdminName       string
	AdminEmail      string
	AdminPassword   string
	SeedDemo        bool
}

// MinJWTSecretLength is the shortest HMAC key accepted for signing access tokens.
const MinJWTSecretLength = 32

// Published defaults must never verify tokens in a real deployment: anyone can
// read them from the repository and forge an admin token.
var knownWeakSecrets = map[string]struct{}{
	"gopress-development-secret-change-me":                  {},
	"local-only-change-this-jwt-secret-32-chars":            {},
	"change-this-to-a-long-random-secret-at-least-32-chars": {},
	"change-this-to-a-long-random-secret-at-least-32-chara": {},
	"replace-with-openssl-rand-hex-32":                      {},
	"changeme":                                              {},
	"secret":                                                {},
}

var knownWeakPasswords = map[string]struct{}{
	"gopress123!":                    {},
	"replace-with-a-strong-password": {},
	"password":                       {},
	"password123!":                   {},
	"admin":                          {},
	"admin123456!":                   {},
}

func Load() (Config, error) {
	cfg := Config{
		Port:          getEnv("PORT", "8080"),
		MySQLDSN:      getEnv("MYSQL_DSN", "gopress:gopress@tcp(127.0.0.1:3306)/gopress?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		// No fallback on purpose: a default signing key makes every deployment
		// with an unset JWT_SECRET trivially forgeable.
		JWTSecret:  os.Getenv("JWT_SECRET"),
		AdminName:  getEnv("ADMIN_NAME", "GoPress Admin"),
		AdminEmail: getEnv("ADMIN_EMAIL", "admin@gopress.dev"),
		// No fallback on purpose: a published default admin password hands over
		// the whole admin console to anyone who reads the repository.
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}

	var err error
	if cfg.RedisDB, err = getInt("REDIS_DB", 0); err != nil {
		return Config{}, err
	}
	if cfg.AccessTokenTTL, err = getDuration("ACCESS_TOKEN_TTL", 15*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.RefreshTokenTTL, err = getDuration("REFRESH_TOKEN_TTL", 168*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.CookieSecure, err = getBool("COOKIE_SECURE", false); err != nil {
		return Config{}, err
	}
	if cfg.SeedDemo, err = getBool("SEED_DEMO", false); err != nil {
		return Config{}, err
	}
	if cfg.AccessTokenTTL <= 0 || cfg.RefreshTokenTTL <= 0 {
		return Config{}, fmt.Errorf("token TTLs must be positive")
	}

	if err := validateSecret(cfg.JWTSecret); err != nil {
		return Config{}, err
	}
	if strings.TrimSpace(cfg.AdminPassword) != "" {
		if err := ValidateAdminPassword(cfg.AdminPassword); err != nil {
			return Config{}, err
		}
	}

	origins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"), ",")
	cfg.CORSOrigins = make([]string, 0, len(origins))
	for _, origin := range origins {
		if origin = strings.TrimSpace(origin); origin != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, origin)
		}
	}

	proxies := strings.Split(os.Getenv("TRUSTED_PROXIES"), ",")
	cfg.TrustedProxies = make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		if proxy = strings.TrimSpace(proxy); proxy != "" {
			cfg.TrustedProxies = append(cfg.TrustedProxies, proxy)
		}
	}

	return cfg, nil
}

func validateSecret(secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return fmt.Errorf("JWT_SECRET is required: generate one with `openssl rand -hex 32`")
	}
	if len(secret) < MinJWTSecretLength {
		return fmt.Errorf("JWT_SECRET must be at least %d characters", MinJWTSecretLength)
	}
	if _, weak := knownWeakSecrets[strings.ToLower(secret)]; weak {
		return fmt.Errorf("JWT_SECRET matches a published default value: generate a unique secret with `openssl rand -hex 32`")
	}
	return nil
}

// ValidateAdminPassword enforces a minimum strength so the seeded administrator
// cannot start life with a guessable password.
func ValidateAdminPassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("ADMIN_PASSWORD must contain at least 12 characters")
	}
	if _, weak := knownWeakPasswords[strings.ToLower(strings.TrimSpace(password))]; weak {
		return fmt.Errorf("ADMIN_PASSWORD is a well-known default: choose a unique password")
	}
	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	classes := 0
	for _, present := range []bool{hasLower, hasUpper, hasDigit, hasSymbol} {
		if present {
			classes++
		}
	}
	if classes < 3 {
		return fmt.Errorf("ADMIN_PASSWORD must mix at least three of: lowercase, uppercase, digits, symbols")
	}
	return nil
}

// ValidateAdminEmail checks the seeding mailbox before it is persisted.
func ValidateAdminEmail(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return fmt.Errorf("ADMIN_EMAIL cannot be empty")
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return fmt.Errorf("ADMIN_EMAIL must be a valid email address")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return parsed, nil
}

func getBool(key string, fallback bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s: %w", key, err)
	}
	return parsed, nil
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return parsed, nil
}
