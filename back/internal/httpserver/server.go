package httpserver

import (
	"net/http"
	"time"

	"test-builder/internal/attempts"
	"test-builder/internal/auth"
	"test-builder/internal/dashboard"
	"test-builder/internal/publications"
	"test-builder/internal/tests"

	_ "test-builder/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(
	authHandler *auth.Handler,
	dashboardHandler *dashboard.Handler,
	testsHandler *tests.Handler,
	publicationsHandler *publications.Handler,
	attemptsHandler *attempts.Handler,
	rdb *redis.Client,
	frontendOrigin string,
	appEnv string,
) *gin.Engine {

	binding.EnableDecoderDisallowUnknownFields = true

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		panic("httpserver: set trusted proxies: " + err.Error())
	}
	isProduction := appEnv == "production"

	router.Use(
		gin.Logger(),
		gin.Recovery(),
		securityHeaders(isProduction),
		cors.New(cors.Config{
			AllowOrigins:     []string{frontendOrigin},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-CSRF-Token"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),

		maxBodySize(1<<20),
	)

	// 404 / 405 в JSON, чтобы фронт не получал plain-text
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Route not found",
			},
		})
	})
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": gin.H{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "Method not allowed",
			},
		})
	})
	authIPLimit := redis_rate.PerMinute(10)
	authRoutes := router.Group("/auth")
	authRoutes.Use(noStore())
	{
		authRoutes.POST("/register", rateLimitByIP(rdb, authIPLimit), authHandler.Register)
		authRoutes.POST("/login", rateLimitByIP(rdb, authIPLimit), authHandler.Login)
		authRoutes.POST("/logout", authHandler.RequireAuth(), authHandler.Logout)
		authRoutes.GET("/me", authHandler.RequireAuth(), authHandler.Me)
	}

	dashboardRoutes := router.Group("/app")
	dashboardRoutes.Use(authHandler.RequireAuth())
	{
		dashboardRoutes.GET("/overview", dashboardHandler.GetOverview)
	}

	testsRoutes := router.Group("/tests")
	testsRoutes.Use(authHandler.RequireAuth())
	{
		testsRoutes.POST("", testsHandler.CreateTest)
		testsRoutes.GET("", testsHandler.ListTests)
		testsRoutes.GET("/:testId", testsHandler.GetTest)
		testsRoutes.PUT("/:testId/draft", testsHandler.SaveDraft)
		testsRoutes.POST("/:testId/validate", publicationsHandler.Validate)
		testsRoutes.POST("/:testId/publish", publicationsHandler.Publish)
		testsRoutes.GET("/:testId/publication", publicationsHandler.GetPublication)
	}

	publicTests := router.Group("/public/tests")
	publicTests.Use(noStore())
	{

		publicTests.GET("/:token",
			rateLimitByIP(rdb, redis_rate.PerMinute(60)),
			attemptsHandler.GetPublicTest,
		)
		publicTests.POST("/:token/attempts",
			rateLimitByIP(rdb, redis_rate.PerMinute(60)),
			rateLimitByParam(rdb, redis_rate.PerMinute(500), "token"),
			attemptsHandler.CreateAttempt,
		)
	}

	publicAttempts := router.Group("/public/attempts")
	publicAttempts.Use(noStore(), attemptsHandler.RequireAttemptCookie())
	{
		publicAttempts.GET("/:attemptId", attemptsHandler.GetState)
		publicAttempts.PUT("/:attemptId/answers", attemptsHandler.SaveAnswers)
		publicAttempts.POST("/:attemptId/submit", attemptsHandler.Submit)
		publicAttempts.GET("/:attemptId/result", attemptsHandler.GetResult)
	}

	if !isProduction {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return router
}

// noStore запрещает кеширование ответов группы /auth.
func noStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

func securityHeaders(isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		if isProduction {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}
func maxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
