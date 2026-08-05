package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiterEntry struct {
	lastRequest time.Time
}

var (
	rateLimiters = make(map[uint]*rateLimiterEntry)
	rlMutex      sync.RWMutex
)

// RateLimitMiddleware limits requests per user.
// interval specifies the minimum time between requests.
func RateLimitMiddleware(interval time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		uid := userID.(uint)

		rlMutex.Lock()
		entry, found := rateLimiters[uid]

		if found && time.Since(entry.lastRequest) < interval {
			rlMutex.Unlock()
			remaining := interval - time.Since(entry.lastRequest)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Too many requests. Please wait before submitting again.",
				"retry_after": remaining.Seconds(),
			})
			c.Abort()
			return
		}

		if !found {
			rateLimiters[uid] = &rateLimiterEntry{}
			entry = rateLimiters[uid]
		}
		entry.lastRequest = time.Now()
		rlMutex.Unlock()

		c.Next()
	}
}
