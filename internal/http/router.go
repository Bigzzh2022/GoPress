package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gopress/internal/config"
	"gopress/internal/http/middleware"
	"gopress/internal/http/response"
	"gopress/internal/modules/auth"
	"gopress/internal/modules/dashboard"
	"gopress/internal/modules/posts"
	"gopress/internal/modules/profile"
	"gopress/internal/modules/taxonomy"
	"gopress/internal/platform/cache"
	"gopress/internal/platform/markdown"
)

func NewRouter(cfg config.Config, db *gorm.DB, publicCache *cache.Cache) *gin.Engine {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORS(cfg.CORSOrigins))

	renderer := markdown.NewRenderer()
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL)
	authHandler := auth.NewHandler(db, tokens, publicCache.Client, cfg.RefreshTokenTTL, cfg.CookieSecure)
	postService := posts.NewService(db, renderer, publicCache)
	postHandler := posts.NewHandler(postService, publicCache, renderer)
	taxonomyService := taxonomy.NewService(db, publicCache)
	taxonomyHandler := taxonomy.NewHandler(taxonomyService, publicCache)
	profileHandler := profile.NewHandler(db, publicCache)
	dashboardHandler := dashboard.NewHandler(db)

	api := router.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})
	api.GET("/posts", postHandler.PublicList)
	api.GET("/posts/:slug", postHandler.PublicDetail)
	api.GET("/posts/:slug/related", postHandler.Related)
	api.GET("/categories", taxonomyHandler.PublicCategories)
	api.GET("/tags", taxonomyHandler.PublicTags)
	api.GET("/profile", profileHandler.Public)

	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/refresh", authHandler.Refresh)
	api.POST("/auth/logout", authHandler.Logout)
	api.GET("/auth/me", authHandler.Authenticate(), authHandler.Me)

	admin := api.Group("/admin", authHandler.Authenticate(), auth.RequireAdmin())
	admin.GET("/dashboard", dashboardHandler.Get)
	admin.GET("/posts", postHandler.AdminList)
	admin.POST("/posts", postHandler.Create)
	admin.GET("/posts/:id", postHandler.AdminDetail)
	admin.PUT("/posts/:id", postHandler.Update)
	admin.DELETE("/posts/:id", postHandler.Delete)
	admin.PATCH("/posts/:id/status", postHandler.SetStatus)
	admin.POST("/markdown/preview", postHandler.Preview)
	admin.GET("/categories", taxonomyHandler.AdminCategories)
	admin.POST("/categories", taxonomyHandler.CreateCategory)
	admin.PUT("/categories/:id", taxonomyHandler.UpdateCategory)
	admin.DELETE("/categories/:id", taxonomyHandler.DeleteCategory)
	admin.GET("/tags", taxonomyHandler.AdminTags)
	admin.POST("/tags", taxonomyHandler.CreateTag)
	admin.PUT("/tags/:id", taxonomyHandler.UpdateTag)
	admin.DELETE("/tags/:id", taxonomyHandler.DeleteTag)
	admin.GET("/profile", profileHandler.Admin)
	admin.PUT("/profile", profileHandler.Update)

	router.NoRoute(func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, "route not found")
	})
	return router
}
