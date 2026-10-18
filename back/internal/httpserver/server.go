package httpserver

import (
	"net/http"
	"time"

	"test-builder/internal/auth"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func New(
	authHandler *auth.Handler,
	frontendOrigin string,
	appEnv string,
) *gin.Engine {
	
	binding.EnableDecoderDisallowUnknownFields = true

	router := gin.New()

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


	//  без префикса /api — nginx его срезает 
	authRoutes := router.Group("/auth")
	authRoutes.Use(noStore())
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/logout", authHandler.Logout)
		authRoutes.GET("/app", authHandler.RequireAuth(), authHandler.Me)
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