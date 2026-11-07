package httpserver

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

//  ограничивает запросы по IP-адресу 
// используется для публичных эндпоинтов и /auth/login 
func rateLimitByIP(rdb *redis.Client, limit redis_rate.Limit) gin.HandlerFunc {
	return rateLimitByKey(rdb, limit, func(c *gin.Context) string {
		return "ip:" + c.ClientIP()
	})
}

// ограничивает по значению URL-параметра 
func rateLimitByParam(rdb *redis.Client, limit redis_rate.Limit, param string) gin.HandlerFunc {
	return rateLimitByKey(rdb, limit, func(c *gin.Context) string {
		v := c.Param(param)
		if v == "" {
			return ""
		}
		return "param:" + param + ":" + v
	})
}

// используется для попыток ученика: каждая попытка имеет свой лимит
func rateLimitByCookie(rdb *redis.Client, limit redis_rate.Limit, cookieName string) gin.HandlerFunc {
	return rateLimitByKey(rdb, limit, func(c *gin.Context) string {
		v, err := c.Cookie(cookieName)
		if err != nil || v == "" {
			return ""
		}
		return "cookie:" + cookieName + ":" + v
	})
}

func rateLimitByKey(rdb *redis.Client, limit redis_rate.Limit, keyFn func(*gin.Context) string) gin.HandlerFunc {
	limiter := redis_rate.NewLimiter(rdb)

	return func(c *gin.Context) {
		key := keyFn(c)
		if key == "" {
			c.Next()
			return
		}

		res, err := limiter.Allow(c.Request.Context(), key, limit)
		if err != nil {
			// потом  метрика 
			slog.ErrorContext(c.Request.Context(), "rate limiter error", "error", err)
			c.Next()
			return
		}

		c.Header("RateLimit-Limit", strconv.Itoa(limit.Burst))
		c.Header("RateLimit-Remaining", strconv.Itoa(res.Remaining))

		if res.Allowed == 0 {
			c.Header("Retry-After", strconv.Itoa(int(res.RetryAfter.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "TOO_MANY_REQUESTS",
					"message": "Too many requests. Please try again later.",
				},
			})
			return
		}

		c.Next()
	}
}