package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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
	CookieSecure    bool
	AdminName       string
	AdminEmail      string
	AdminPassword   string
	SeedDemo        bool
}

func Load() (Config, error) {
	cfg := Config{
		Port:          getEnv("PORT", "8080"),
		MySQLDSN:      getEnv("MYSQL_DSN", "gopress:gopress@tcp(127.0.0.1:3306)/gopress?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		JWTSecret:     getEnv("JWT_SECRET", "gopress-development-secret-change-me"),
		CookieSecure:  false,
		AdminName:     getEnv("ADMIN_NAME", "GoPress Admin"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@gopress.dev"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "GoPress123!"),
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
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return Config{}, fmt.Errorf("JWT_SECRET cannot be empty")
	}

	origins := strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173"), ",")
	cfg.CORSOrigins = make([]string, 0, len(origins))
	for _, origin := range origins {
		if origin = strings.TrimSpace(origin); origin != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, origin)
		}
	}
	return cfg, nil
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
