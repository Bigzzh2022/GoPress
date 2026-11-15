package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gopress/internal/http/response"
	"gopress/internal/model"
)

const (
	userIDKey   = "auth_user_id"
	userRoleKey = "auth_user_role"
	cookieName  = "refresh_token"
	cookiePath  = "/api/v1/auth"
)

type Handler struct {
	db         *gorm.DB
	tokens     *TokenManager
	redis      *redis.Client
	refreshTTL time.Duration
	secure     bool
}

type session struct {
	UserID uint `json:"user_id"`
}

type loginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type authOutput struct {
	AccessToken string        `json:"access_token"`
	User        model.UserDTO `json:"user"`
}

func NewHandler(db *gorm.DB, tokens *TokenManager, redisClient *redis.Client, refreshTTL time.Duration, secure bool) *Handler {
	return &Handler{db: db, tokens: tokens, redis: redisClient, refreshTTL: refreshTTL, secure: secure}
}

func CurrentUserID(c *gin.Context) uint {
	value, _ := c.Get(userIDKey)
	id, _ := value.(uint)
	return id
}

func (h *Handler) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "missing bearer token")
			return
		}
		id, role, err := h.tokens.Parse(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "invalid or expired access token")
			return
		}
		c.Set(userIDKey, id)
		c.Set(userRoleKey, role)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(userRoleKey)
		if role != model.RoleAdmin {
			response.Error(c, http.StatusForbidden, response.CodeForbidden, "admin access required")
			return
		}
		c.Next()
	}
}

func (h *Handler) Login(c *gin.Context) {
	var input loginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid login input")
		return
	}
	if !h.allowLogin(c.Request.Context(), c.ClientIP()) {
		response.Error(c, http.StatusTooManyRequests, response.CodeRateLimited, "too many login attempts")
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	var user model.User
	if err := h.db.WithContext(c).Where("email = ?", email).First(&user).Error; err != nil {
		h.invalidCredentials(c)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		h.invalidCredentials(c)
		return
	}
	accessToken, err := h.tokens.Generate(user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "could not issue access token")
		return
	}
	if h.redis != nil {
		_ = h.redis.Del(c, loginRateKey(c.ClientIP())).Err()
	}

	refreshToken, err := newOpaqueToken()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "could not issue refresh token")
		return
	}
	if err := h.storeSession(c, refreshToken, user.ID); err != nil {
		log.Printf("store refresh session: %v", err)
	} else {
		h.setRefreshCookie(c, refreshToken, int(h.refreshTTL.Seconds()), time.Now().Add(h.refreshTTL))
	}
	response.OK(c, authOutput{AccessToken: accessToken, User: model.ToUserDTO(user)})
}

func (h *Handler) Refresh(c *gin.Context) {
	token, err := c.Cookie(cookieName)
	if err != nil || token == "" {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "missing refresh token")
		return
	}
	if h.redis == nil {
		response.Error(c, http.StatusServiceUnavailable, response.CodeUnavailable, "refresh sessions unavailable")
		return
	}
	key := refreshKey(token)
	value, err := h.redis.GetDel(c, key).Bytes()
	if errors.Is(err, redis.Nil) {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "invalid refresh token")
		return
	}
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, response.CodeUnavailable, "refresh sessions unavailable")
		return
	}
	var saved session
	if json.Unmarshal(value, &saved) != nil || saved.UserID == 0 {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "invalid refresh session")
		return
	}
	var user model.User
	if err := h.db.WithContext(c).First(&user, saved.UserID).Error; err != nil {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "session user not found")
		return
	}
	newToken, err := newOpaqueToken()
	if err != nil || h.storeSession(c, newToken, user.ID) != nil {
		response.Error(c, http.StatusServiceUnavailable, response.CodeUnavailable, "could not rotate refresh token")
		return
	}
	accessToken, err := h.tokens.Generate(user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "could not issue access token")
		return
	}
	h.setRefreshCookie(c, newToken, int(h.refreshTTL.Seconds()), time.Now().Add(h.refreshTTL))
	response.OK(c, authOutput{AccessToken: accessToken, User: model.ToUserDTO(user)})
}

func (h *Handler) Logout(c *gin.Context) {
	if token, err := c.Cookie(cookieName); err == nil && token != "" && h.redis != nil {
		_ = h.redis.Del(c, refreshKey(token)).Err()
	}
	h.setRefreshCookie(c, "", -1, time.Unix(1, 0))
	response.OK(c, gin.H{"logged_out": true})
}

func (h *Handler) Me(c *gin.Context) {
	var user model.User
	if err := h.db.WithContext(c).First(&user, CurrentUserID(c)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, "user not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "could not load user")
		return
	}
	response.OK(c, model.ToUserDTO(user))
}

func (h *Handler) invalidCredentials(c *gin.Context) {
	response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "invalid email or password")
}

func (h *Handler) allowLogin(ctx context.Context, ip string) bool {
	if h.redis == nil {
		return true
	}
	key := loginRateKey(ip)
	count, err := h.redis.Incr(ctx, key).Result()
	if err != nil {
		return true
	}
	if count == 1 {
		if err := h.redis.Expire(ctx, key, time.Minute).Err(); err != nil {
			_ = h.redis.Del(ctx, key).Err()
			return true
		}
	}
	return count <= 10
}

func loginRateKey(ip string) string {
	digest := sha256.Sum256([]byte(ip))
	return "gopress:login:" + hex.EncodeToString(digest[:])
}

func newOpaqueToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func refreshKey(token string) string {
	digest := sha256.Sum256([]byte(token))
	return "gopress:refresh:" + hex.EncodeToString(digest[:])
}

func (h *Handler) storeSession(ctx context.Context, token string, userID uint) error {
	if h.redis == nil {
		return errors.New("redis unavailable")
	}
	value, err := json.Marshal(session{UserID: userID})
	if err != nil {
		return err
	}
	return h.redis.Set(ctx, refreshKey(token), value, h.refreshTTL).Err()
}

func (h *Handler) setRefreshCookie(c *gin.Context, value string, maxAge int, expires time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: cookieName, Value: value, Path: cookiePath, MaxAge: maxAge,
		Expires: expires, HttpOnly: true, Secure: h.secure, SameSite: http.SameSiteLaxMode,
	})
}
