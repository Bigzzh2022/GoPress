package taxonomy

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gopress/internal/http/response"
	"gopress/internal/model"
	"gopress/internal/platform/cache"
)

type Handler struct {
	service *Service
	cache   *cache.Cache
}

func NewHandler(service *Service, publicCache *cache.Cache) *Handler {
	return &Handler{service: service, cache: publicCache}
}

func (h *Handler) PublicCategories(c *gin.Context) {
	key := h.cache.PublicKey(c, "categories", "all")
	var cached []model.CategoryDTO
	if h.cache.GetJSON(c, key, &cached) {
		response.OK(c, cached)
		return
	}
	items, err := h.service.ListCategories(c, true)
	if err != nil {
		handleError(c, err)
		return
	}
	h.cache.SetJSON(c, key, items)
	response.OK(c, items)
}

func (h *Handler) AdminCategories(c *gin.Context) {
	items, err := h.service.ListCategories(c, false)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *Handler) PublicTags(c *gin.Context) {
	key := h.cache.PublicKey(c, "tags", "all")
	var cached []model.TagDTO
	if h.cache.GetJSON(c, key, &cached) {
		response.OK(c, cached)
		return
	}
	items, err := h.service.ListTags(c, true)
	if err != nil {
		handleError(c, err)
		return
	}
	h.cache.SetJSON(c, key, items)
	response.OK(c, items)
}

func (h *Handler) AdminTags(c *gin.Context) {
	items, err := h.service.ListTags(c, false)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var input CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid category input")
		return
	}
	item, err := h.service.CreateCategory(c, input)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, ok := taxonomyID(c)
	if !ok {
		return
	}
	var input CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid category input")
		return
	}
	item, err := h.service.UpdateCategory(c, id, input)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, ok := taxonomyID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteCategory(c, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) CreateTag(c *gin.Context) {
	var input TagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid tag input")
		return
	}
	item, err := h.service.CreateTag(c, input)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *Handler) UpdateTag(c *gin.Context) {
	id, ok := taxonomyID(c)
	if !ok {
		return
	}
	var input TagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid tag input")
		return
	}
	item, err := h.service.UpdateTag(c, id, input)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *Handler) DeleteTag(c *gin.Context) {
	id, ok := taxonomyID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteTag(c, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func taxonomyID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid resource id")
		return 0, false
	}
	return uint(value), true
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrCategoryNotFound), errors.Is(err, ErrTagNotFound):
		response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
	case errors.Is(err, ErrConflict):
		response.Error(c, http.StatusConflict, response.CodeConflict, "slug already exists")
	case errors.Is(err, ErrInvalid):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid taxonomy input")
	default:
		log.Printf("taxonomy operation: %v", err)
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "taxonomy operation failed")
	}
}
