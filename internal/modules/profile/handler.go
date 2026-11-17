package profile

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gopress/internal/http/response"
	"gopress/internal/model"
	"gopress/internal/platform/cache"
	"gopress/internal/platform/database"
)

type Handler struct {
	db    *gorm.DB
	cache *cache.Cache
}

type UpdateInput struct {
	Name   string `json:"name" binding:"required,max=100"`
	Email  string `json:"email" binding:"required,email,max=191"`
	Avatar string `json:"avatar" binding:"omitempty,max=500"`
	Bio    string `json:"bio"`
}

func NewHandler(db *gorm.DB, publicCache *cache.Cache) *Handler {
	return &Handler{db: db, cache: publicCache}
}

func (h *Handler) Public(c *gin.Context) {
	key := h.cache.PublicKey(c, "profile", "site")
	var cached model.PublicUserDTO
	if h.cache.GetJSON(c, key, &cached) {
		response.OK(c, cached)
		return
	}
	user, err := h.siteUser(c)
	if err != nil {
		h.handleError(c, err)
		return
	}
	// The public profile deliberately omits the email address; it is the same
	// mailbox that signs in to the admin console.
	dto := model.ToPublicUserDTO(user)
	h.cache.SetJSON(c, key, dto)
	response.OK(c, dto)
}

func (h *Handler) Admin(c *gin.Context) {
	user, err := h.siteUser(c)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, model.ToUserDTO(user))
}

func (h *Handler) Update(c *gin.Context) {
	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid profile input")
		return
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "profile name cannot be empty")
		return
	}
	user, err := h.siteUser(c)
	if err != nil {
		h.handleError(c, err)
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	var count int64
	if err := h.db.WithContext(c).Model(&model.User{}).
		Where("email = ? AND id <> ?", email, user.ID).Count(&count).Error; err != nil {
		h.handleError(c, err)
		return
	}
	if count > 0 {
		response.Error(c, http.StatusConflict, response.CodeConflict, "email already exists")
		return
	}
	if err := h.db.WithContext(c).Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"name": name, "email": email, "avatar": strings.TrimSpace(input.Avatar), "bio": input.Bio,
	}).Error; err != nil {
		if database.IsDuplicate(err) {
			response.Error(c, http.StatusConflict, response.CodeConflict, "email already exists")
			return
		}
		h.handleError(c, err)
		return
	}
	_ = h.cache.BumpVersion(c)
	user, err = h.siteUser(c)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, model.ToUserDTO(user))
}

func (h *Handler) siteUser(c *gin.Context) (model.User, error) {
	var user model.User
	err := h.db.WithContext(c).Where("role = ?", model.RoleAdmin).Order("id ASC").First(&user).Error
	return user, err
}

func (h *Handler) handleError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "profile not found")
		return
	}
	response.Error(c, http.StatusInternalServerError, response.CodeInternal, "profile operation failed")
}
