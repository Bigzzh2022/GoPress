package httpapi

import (
	"net/http"
	"time"

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

const (
	maxRequestBodyBytes = 2 << 20 // 2 MiB, matching the reverse proxy limit
	publicRateLimit     = 120     // requests per minute per client IP
	loginRateLimit      = 10      // failed attempts per minute per client IP
	refreshRateLimit    = 30      // refresh calls per minute per client IP
)

func NewRouter(cfg config.Config, db *gorm.DB, publicCache *cache.Cache) *gin.Engine {
	router := gin.New()
	// Only trust forwarding headers from explicit proxies; otherwise ClientIP
	// would collapse to the proxy address and turn per-IP limits global.
	var trustedProxies []string
	if len(cfg.TrustedProxies) > 0 {
		trustedProxies = cfg.TrustedProxies
	}
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		panic("invalid TRUSTED_PROXIES: " + err.Error())
	}
	router.Use(
		middleware.RequestLogger(),
		gin.Recovery(),
		middleware.SecurityHeaders(),
		middleware.CORS(cfg.CORSOrigins),
	)

	renderer := markdown.NewRenderer()
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL)
	loginLimiter := middleware.NewRateLimiter(publicCache.Client, "login", loginRateLimit, time.Minute)
	refreshLimiter := middleware.NewRateLimiter(publicCache.Client, "refresh", refreshRateLimit, time.Minute)
	publicLimiter := middleware.NewRateLimiter(publicCache.Client, "public", publicRateLimit, time.Minute)
	authHandler := auth.NewHandler(
		db, tokens, publicCache.Client, cfg.RefreshTokenTTL, cfg.CookieSecure,
		loginLimiter, refreshLimiter,
	)
	postService := posts.NewService(db, renderer, publicCache)
	postHandler := posts.NewHandler(postService, publicCache, renderer)
	taxonomyService := taxonomy.NewService(db, publicCache)
	taxonomyHandler := taxonomy.NewHandler(taxonomyService, publicCache)
	profileHandler := profile.NewHandler(db, publicCache)
	dashboardHandler := dashboard.NewHandler(db)

	api := router.Group("/api/v1")
	api.Use(middleware.BodyLimit(maxRequestBodyBytes))

	api.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	public := api.Group("")
	public.Use(middleware.Limit(publicLimiter))
	public.GET("/posts", postHandler.PublicList)
	public.GET("/posts/:slug", postHandler.PublicDetail)
	public.GET("/posts/:slug/related", postHandler.Related)
	public.GET("/categories", taxonomyHandler.PublicCategories)
	public.GET("/tags", taxonomyHandler.PublicTags)
	public.GET("/profile", profileHandler.Public)

	api.POST("/auth/login", authHandler.Login)
	trustedOrigin := middleware.RequireTrustedOrigin(cfg.CORSOrigins)
	api.POST("/auth/refresh", trustedOrigin, authHandler.Refresh)
	api.POST("/auth/logout", trustedOrigin, authHandler.Logout)
	api.GET("/auth/me", authHandler.Authenticate(), authHandler.Me)

	admin := api.Group("/admin", authHandler.Authenticate(), auth.RequireAdmin(db))
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
