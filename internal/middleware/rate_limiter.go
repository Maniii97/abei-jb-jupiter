package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis *redis.Client
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{redis: redis}
}

// RateLimit middleware limits requests per IP/user
func (rl *RateLimiter) RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Using IP address as the key for rate limiting
		key := fmt.Sprintf("rate_limit:%s", c.ClientIP())

		ctx := c.Request.Context()

		// Get current count
		current, err := rl.redis.Get(ctx, key).Int()
		if err == redis.Nil {
			// First request, set counter
			err = rl.redis.Set(ctx, key, 1, window).Err()
			if err != nil {
				// If Redis fails, allow the request (fail open)
				c.Next()
				return
			}
			c.Next()
			return
		} else if err != nil {
			// Redis error, allow request (fail open)
			c.Next()
			return
		}

		// Check if limit exceeded
		if current >= requests {
			// Get TTL for rate limit reset time
			ttl, _ := rl.redis.TTL(ctx, key).Result()

			c.Header("X-Rate-Limit-Limit", strconv.Itoa(requests))
			c.Header("X-Rate-Limit-Remaining", "0")
			c.Header("X-Rate-Limit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": int(ttl.Seconds()),
			})
			c.Abort()
			return
		}

		// Increment counter
		newCount, err := rl.redis.Incr(ctx, key).Result()
		if err != nil {
			// Redis error, allow request (fail open)
			c.Next()
			return
		}

		// Set headers
		remaining := requests - int(newCount)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-Rate-Limit-Limit", strconv.Itoa(requests))
		c.Header("X-Rate-Limit-Remaining", strconv.Itoa(remaining))

		c.Next()
	}
}

// UserRateLimit uses authenticated user ID instead of IP
func (rl *RateLimiter) UserRateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// No user ID, fall back to IP-based rate limiting
			rl.RateLimit(requests, window)(c)
			return
		}

		key := fmt.Sprintf("rate_limit:user:%v", userID)

		ctx := c.Request.Context()

		// Get current count
		current, err := rl.redis.Get(ctx, key).Int()
		if err == redis.Nil {
			// First request, set counter
			err = rl.redis.Set(ctx, key, 1, window).Err()
			if err != nil {
				// If Redis fails, allow the request (fail open)
				c.Next()
				return
			}
			c.Next()
			return
		} else if err != nil {
			// Redis error, allow request (fail open)
			c.Next()
			return
		}

		// Check if limit exceeded
		if current >= requests {
			// Get TTL for rate limit reset time
			ttl, _ := rl.redis.TTL(ctx, key).Result()

			c.Header("X-Rate-Limit-Limit", strconv.Itoa(requests))
			c.Header("X-Rate-Limit-Remaining", "0")
			c.Header("X-Rate-Limit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": int(ttl.Seconds()),
			})
			c.Abort()
			return
		}

		// Increment counter
		newCount, err := rl.redis.Incr(ctx, key).Result()
		if err != nil {
			// Redis error, allow request (fail open)
			c.Next()
			return
		}

		// Set headers
		remaining := requests - int(newCount)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-Rate-Limit-Limit", strconv.Itoa(requests))
		c.Header("X-Rate-Limit-Remaining", strconv.Itoa(remaining))

		c.Next()
	}
}
