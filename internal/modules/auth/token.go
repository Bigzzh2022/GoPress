package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gopress/internal/model"
)

const (
	issuer   = "gopress"
	audience = "gopress-api"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Identity is the authenticated principal decoded from an access token.
type Identity struct {
	UserID    uint
	Role      string
	JTI       string
	ExpiresAt time.Time
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

func (m *TokenManager) Generate(user model.User) (string, error) {
	now := time.Now()
	// A unique token ID lets a single access token be revoked on logout instead
	// of staying valid until it expires.
	jti, err := newOpaqueToken()
	if err != nil {
		return "", err
	}
	claims := Claims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:       jti,
			Issuer:   issuer,
			Audience: jwt.ClaimStrings{audience},
			Subject:  strconv.FormatUint(uint64(user.ID), 10),
			IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *TokenManager) Parse(value string) (Identity, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(value, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer), jwt.WithAudience(audience), jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return Identity{}, errors.New("invalid access token")
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || id == 0 {
		return Identity{}, errors.New("invalid token subject")
	}
	if claims.ID == "" || claims.ExpiresAt == nil {
		return Identity{}, errors.New("invalid access token claims")
	}
	return Identity{
		UserID:    uint(id),
		Role:      claims.Role,
		JTI:       claims.ID,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
