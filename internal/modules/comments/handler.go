package comments

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"gopress/internal/http/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createInput struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	QQ       string `json:"qq"`
	Content  string `json:"content"`
	ParentID *uint  `json:"parent_id"`
	Website  string `json:"website"`
}

type publicListQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1,max=1000"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type adminListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1,max=200"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty,max=200"`
	PostID   uint   `form:"post_id"`
}

func (h *Handler) PublicList(c *gin.Context) {
	var query publicListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid comment query")
		return
	}
	result, err := h.service.List(c, c.Param("slug"), ListParams{
		Page: query.Page, PageSize: query.PageSize,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Create(c *gin.Context) {
	var input createInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "评论内容格式不正确")
		return
	}
	if strings.TrimSpace(input.Website) != "" {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "评论提交失败，请稍后再试")
		return
	}
	comment, err := h.service.Create(c, c.Param("slug"), CreateInput{
		Nickname: input.Nickname, Email: input.Email, QQ: input.QQ,
		Content: input.Content, ParentID: input.ParentID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, comment)
}

func (h *Handler) AdminList(c *gin.Context) {
	var query adminListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid comment query")
		return
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	var postID *uint
	if query.PostID > 0 {
		postID = &query.PostID
	}
	items, total, err := h.service.AdminList(c, AdminListParams{
		Page: query.Page, PageSize: query.PageSize,
		Keyword: query.Keyword, PostID: postID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, response.NewPage(items, total, query.Page, query.PageSize))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid comment id")
		return
	}
	if err := h.service.Delete(c, uint(id)); err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	var validation *ValidationError
	switch {
	case errors.As(err, &validation):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, validation.Message)
	case errors.Is(err, ErrPostNotFound):
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "文章不存在或尚未发布")
	case errors.Is(err, ErrParentNotFound):
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "回复的评论不存在")
	case errors.Is(err, ErrCommentNotFound):
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "评论不存在")
	default:
		log.Printf("comment operation: %v", err)
		response.Error(c, http.StatusInternalServerError, response.CodeInternal, "评论操作失败")
	}
}
