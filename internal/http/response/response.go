package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeBadRequest   = "BAD_REQUEST"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeNotFound     = "NOT_FOUND"
	CodeConflict     = "CONFLICT"
	CodeRateLimited  = "RATE_LIMITED"
	CodeUnavailable  = "SERVICE_UNAVAILABLE"
	CodeInternal     = "INTERNAL_ERROR"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Page[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

func Data(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func OK(c *gin.Context, data any) {
	Data(c, http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	Data(c, http.StatusCreated, data)
}

func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": errorBody{Code: code, Message: message}})
}

func NewPage[T any](items []T, total int64, page, pageSize int) Page[T] {
	if items == nil {
		items = make([]T, 0)
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return Page[T]{Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages}
}
