package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gopress/internal/http/response"
	"gopress/internal/model"
)

type Handler struct {
	db *gorm.DB
}

type Output struct {
	PostCount      int64               `json:"post_count"`
	PublishedCount int64               `json:"published_count"`
	DraftCount     int64               `json:"draft_count"`
	CategoryCount  int64               `json:"category_count"`
	TagCount       int64               `json:"tag_count"`
	RecentPosts    []model.PostSummary `json:"recent_posts"`
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Get(c *gin.Context) {
	var output Output
	queries := []struct {
		destination *int64
		query       *gorm.DB
	}{
		{&output.PostCount, h.db.WithContext(c).Model(&model.Post{})},
		{&output.PublishedCount, h.db.WithContext(c).Model(&model.Post{}).Where("status = ?", model.StatusPublished)},
		{&output.DraftCount, h.db.WithContext(c).Model(&model.Post{}).Where("status = ?", model.StatusDraft)},
		{&output.CategoryCount, h.db.WithContext(c).Model(&model.Category{})},
		{&output.TagCount, h.db.WithContext(c).Model(&model.Tag{})},
	}
	for _, item := range queries {
		if err := item.query.Count(item.destination).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, response.CodeInternal, "could not load dashboard")
			return
		}
	}
	posts := make([]model.Post, 0, 5)
	if err := h.db.WithContext(c).Preload("Author").Preload("Category").Preload("Tags").
		Order("created_at DESC").Order("id DESC").Limit(5).Find(&posts).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "could not load recent posts")
		return
	}
	output.RecentPosts = make([]model.PostSummary, 0, len(posts))
	for _, post := range posts {
		output.RecentPosts = append(output.RecentPosts, model.ToPostSummary(post))
	}
	response.OK(c, output)
}
