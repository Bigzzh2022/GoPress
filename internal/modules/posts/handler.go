package posts

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"gopress/internal/http/response"
	"gopress/internal/model"
	"gopress/internal/modules/auth"
	"gopress/internal/platform/cache"
	"gopress/internal/platform/markdown"
)

type Handler struct {
	service  *Service
	cache    *cache.Cache
	markdown *markdown.Renderer
}

type listQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1,max=200"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty,max=200"`
	Category string `form:"category" binding:"omitempty,max=191"`
	Tag      string `form:"tag" binding:"omitempty,max=191"`
	Status   string `form:"status" binding:"omitempty,oneof=draft published trash"`
}

type statusInput struct {
	Status string `json:"status" binding:"required,oneof=draft published"`
}

type previewInput struct {
	Content string `json:"content" binding:"max=262144"`
}

func NewHandler(service *Service, publicCache *cache.Cache, renderer *markdown.Renderer) *Handler {
	return &Handler{service: service, cache: publicCache, markdown: renderer}
}

func (h *Handler) PublicList(c *gin.Context) {
	query, ok := bindListQuery(c)
	if !ok {
		return
	}
	params := query.params()
	key := h.cache.PublicKey(c, "posts:list", listCacheParameters(query))
	var cached response.Page[model.PostSummary]
	if h.cache.GetJSON(c, key, &cached) {
		response.OK(c, cached)
		return
	}
	items, total, err := h.service.List(c, params, true)
	if err != nil {
		handleError(c, err)
		return
	}
	page := response.NewPage(toSummaries(items), total, params.Page, params.PageSize)
	h.cache.SetJSON(c, key, page)
	response.OK(c, page)
}

func (h *Handler) AdminList(c *gin.Context) {
	query, ok := bindListQuery(c)
	if !ok {
		return
	}
	params := query.params()
	items, total, err := h.service.List(c, params, false)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, response.NewPage(toSummaries(items), total, params.Page, params.PageSize))
}

func (h *Handler) PublicDetail(c *gin.Context) {
	slug := c.Param("slug")
	key := h.cache.PublicKey(c, "posts:detail", slug)
	var cached model.PostDetail
	if h.cache.GetJSON(c, key, &cached) {
		response.OK(c, cached)
		return
	}
	post, err := h.service.PublicDetail(c, slug)
	if err != nil {
		handleError(c, err)
		return
	}
	detail := model.ToPostDetail(post)
	h.cache.SetJSON(c, key, detail)
	response.OK(c, detail)
}

func (h *Handler) AdminDetail(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	post, err := h.service.AdminDetail(c, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, model.ToPostDetail(post))
}

func (h *Handler) Related(c *gin.Context) {
	slug := c.Param("slug")
	key := h.cache.PublicKey(c, "posts:related", slug)
	var cached []model.PostSummary
	if h.cache.GetJSON(c, key, &cached) {
		response.OK(c, cached)
		return
	}
	posts, err := h.service.Related(c, slug)
	if err != nil {
		handleError(c, err)
		return
	}
	items := toSummaries(posts)
	h.cache.SetJSON(c, key, items)
	response.OK(c, items)
}

func (h *Handler) Create(c *gin.Context) {
	var input CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid post input")
		return
	}
	post, err := h.service.Create(c, auth.CurrentUserID(c), input)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, model.ToPostDetail(post))
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid post input")
		return
	}
	post, err := h.service.Update(c, id, input)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, model.ToPostDetail(post))
}

func (h *Handler) SetStatus(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var input statusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid status")
		return
	}
	post, err := h.service.SetStatus(c, id, input.Status)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, model.ToPostDetail(post))
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) Restore(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	post, err := h.service.Restore(c, id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, model.ToPostDetail(post))
}

func (h *Handler) Purge(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.service.Purge(c, id); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) Batch(c *gin.Context) {
	var input BatchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid batch input")
		return
	}
	affected, err := h.service.Batch(c, input)
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"affected": affected})
}

func (h *Handler) Preview(c *gin.Context) {
	var input previewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid markdown input")
		return
	}
	html, err := h.markdown.Render(input.Content)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "could not render markdown")
		return
	}
	response.OK(c, gin.H{"html": html})
}

func bindListQuery(c *gin.Context) (listQuery, bool) {
	var query listQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid pagination or filters")
		return query, false
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 10
	}
	return query, true
}

func (q listQuery) params() ListParams {
	return ListParams{
		Page: q.Page, PageSize: q.PageSize, Keyword: q.Keyword,
		Category: q.Category, Tag: q.Tag, Status: q.Status,
	}
}

func listCacheParameters(query listQuery) string {
	values := url.Values{}
	values.Set("page", strconv.Itoa(query.Page))
	values.Set("page_size", strconv.Itoa(query.PageSize))
	values.Set("keyword", query.Keyword)
	values.Set("category", query.Category)
	values.Set("tag", query.Tag)
	return values.Encode()
}

func toSummaries(posts []model.Post) []model.PostSummary {
	items := make([]model.PostSummary, 0, len(posts))
	for _, post := range posts {
		items = append(items, model.ToPostSummary(post))
	}
	return items
}

func pathID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid resource id")
		return 0, false
	}
	return uint(value), true
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "post not found")
	case errors.Is(err, ErrConflict):
		response.Error(c, http.StatusConflict, response.CodeConflict, "post slug already exists")
	case errors.Is(err, ErrInvalidCategory), errors.Is(err, ErrInvalidTags):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, err.Error())
	case errors.Is(err, ErrNotTrashed):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "article must be in the recycle bin before permanent deletion")
	case errors.Is(err, ErrInvalid):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid post input")
	default:
		log.Printf("post operation: %v", err)
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "post operation failed")
	}
}
