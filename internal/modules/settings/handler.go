package settings

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gopress/internal/http/response"
	"gopress/internal/model"
	"gopress/internal/platform/cache"
)

const (
	defaultSiteName        = "GoPress"
	defaultSiteTagline     = "CODE & CURIOSITY"
	defaultSiteDescription = "关于 Go、分布式系统与工程实践的独立博客。不追逐喧嚣，只记录值得反复阅读的技术思考。"
	defaultFooterText      = "BUILT WITH GO · DESIGNED FOR READING"
	defaultContactEmail    = "hello@gopress.dev"
)

type Handler struct {
	db    *gorm.DB
	cache *cache.Cache
}

type UpdateInput struct {
	SiteName        string `json:"site_name" binding:"required,max=100"`
	SiteTagline     string `json:"site_tagline" binding:"omitempty,max=200"`
	SiteDescription string `json:"site_description" binding:"omitempty,max=1000"`
	FooterText      string `json:"footer_text" binding:"omitempty,max=255"`
	ContactEmail    string `json:"contact_email" binding:"omitempty,email,max=191"`
}

func NewHandler(db *gorm.DB, publicCache *cache.Cache) *Handler {
	return &Handler{db: db, cache: publicCache}
}

func (h *Handler) Public(c *gin.Context) {
	key := h.cache.PublicKey(c, "settings", "site")
	var cached model.SiteSettingsDTO
	if h.cache.GetJSON(c, key, &cached) {
		response.OK(c, cached)
		return
	}
	setting, err := h.load(c)
	if err != nil {
		h.handleError(c, err)
		return
	}
	dto := model.ToSiteSettingsDTO(setting)
	h.cache.SetJSON(c, key, dto)
	response.OK(c, dto)
}

func (h *Handler) Admin(c *gin.Context) {
	setting, err := h.load(c)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, model.ToSiteSettingsDTO(setting))
}

func (h *Handler) Update(c *gin.Context) {
	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid site settings input")
		return
	}
	name := strings.TrimSpace(input.SiteName)
	if name == "" {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "site name cannot be empty")
		return
	}
	setting, err := h.load(c)
	if err != nil {
		h.handleError(c, err)
		return
	}
	updates := map[string]any{
		"site_name":        name,
		"site_tagline":     strings.TrimSpace(input.SiteTagline),
		"site_description": strings.TrimSpace(input.SiteDescription),
		"footer_text":      strings.TrimSpace(input.FooterText),
		"contact_email":    strings.ToLower(strings.TrimSpace(input.ContactEmail)),
	}
	if err := h.db.WithContext(c).Model(&model.SiteSetting{}).Where("id = ?", setting.ID).
		Updates(updates).Error; err != nil {
		h.handleError(c, err)
		return
	}
	_ = h.cache.BumpVersion(c)
	if err := h.db.WithContext(c).First(&setting, setting.ID).Error; err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, model.ToSiteSettingsDTO(setting))
}

func (h *Handler) load(c *gin.Context) (model.SiteSetting, error) {
	defaults := model.SiteSetting{
		ID: 1, SiteName: defaultSiteName, SiteTagline: defaultSiteTagline,
		SiteDescription: defaultSiteDescription, FooterText: defaultFooterText,
		ContactEmail: defaultContactEmail,
	}
	var setting model.SiteSetting
	err := h.db.WithContext(c).Where("id = ?", 1).Attrs(defaults).FirstOrCreate(&setting).Error
	return setting, err
}

func (h *Handler) handleError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "site settings not found")
		return
	}
	log.Printf("site settings operation: %v", err)
	response.Error(c, http.StatusInternalServerError, response.CodeInternal, "site settings operation failed")
}
