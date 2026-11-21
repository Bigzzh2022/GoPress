package updates

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"gopress/internal/http/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Check reports whether the deployed build is behind the configured branch.
func (h *Handler) Check(c *gin.Context) {
	status, err := h.service.Check(c.Request.Context(), c.Query("force") == "true")
	if err != nil {
		var upstream *UpstreamError
		if errors.As(err, &upstream) {
			log.Printf("update check: %v", err)
			if upstream.Status == http.StatusForbidden || upstream.Status == http.StatusTooManyRequests {
				response.Error(c, http.StatusServiceUnavailable, response.CodeRateLimited,
					"GitHub 接口访问受限，请稍后重试或在服务端配置 GITHUB_TOKEN")
				return
			}
			response.Error(c, http.StatusBadGateway, response.CodeUnavailable, "无法获取 GitHub 上的最新版本")
			return
		}
		log.Printf("update check: %v", err)
		response.Error(c, http.StatusBadGateway, response.CodeUnavailable, "无法连接 GitHub，请检查服务器网络")
		return
	}
	response.OK(c, status)
}
