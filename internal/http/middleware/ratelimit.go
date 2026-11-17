package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"gopress/internal/http/response"
)

// maxMemoryKeys bounds the fallback store so a Redis outage cannot turn the
// in-process limiter into an unbounded memory leak.
const maxMemoryKeys = 10000

type rateEntry struct {
	count   int
	resetAt time.Time
}

// RateLimiter counts requests per key inside a fixed window. Redis provides the
// shared counter; an in-process counter keeps throttling active when Redis is
// unavailable instead of failing open.
type RateLimiter struct {
	redis  *redis.Client
	name   string
	limit  int
	window time.Duration

	mu      sync.Mutex
	entries map[string]*rateEntry
}

func NewRateLimiter(client *redis.Client, name string, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:   client,
		name:    name,
		limit:   limit,
		window:  window,
		entries: make(map[string]*rateEntry),
	}
}

// Allow reports whether the key may proceed and, when throttled, how long the
// caller should wait before retrying.
func (l *RateLimiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	if l.redis != nil {
		redisKey := l.redisKey(key)
		count, err := l.redis.Incr(ctx, redisKey).Result()
		if err == nil {
			if count == 1 {
				if expireErr := l.redis.Expire(ctx, redisKey, l.window).Err(); expireErr != nil {
					_ = l.redis.Del(ctx, redisKey).Err()
					return l.allowMemory(key)
				}
			}
			if count > int64(l.limit) {
				ttl, ttlErr := l.redis.TTL(ctx, redisKey).Result()
				if ttlErr != nil || ttl <= 0 {
					ttl = l.window
				}
				return false, ttl
			}
			return true, 0
		}
	}
	return l.allowMemory(key)
}

// Reset clears the counter, e.g. after a successful login.
func (l *RateLimiter) Reset(ctx context.Context, key string) {
	l.mu.Lock()
	delete(l.entries, key)
	l.mu.Unlock()
	if l.redis != nil {
		_ = l.redis.Del(ctx, l.redisKey(key)).Err()
	}
}

// Limit throttles requests by client IP and answers with Retry-After.
func Limit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, retryAfter := limiter.Allow(c.Request.Context(), c.ClientIP())
		if !allowed {
			seconds := int(math.Ceil(retryAfter.Seconds()))
			if seconds < 1 {
				seconds = 1
			}
			c.Header("Retry-After", strconv.Itoa(seconds))
			response.Error(c, http.StatusTooManyRequests, response.CodeRateLimited, "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (l *RateLimiter) allowMemory(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.pruneLocked(now)
	entry, ok := l.entries[key]
	if !ok || now.After(entry.resetAt) {
		l.entries[key] = &rateEntry{count: 1, resetAt: now.Add(l.window)}
		return true, 0
	}
	entry.count++
	if entry.count > l.limit {
		return false, time.Until(entry.resetAt)
	}
	return true, 0
}

func (l *RateLimiter) pruneLocked(now time.Time) {
	for key, entry := range l.entries {
		if now.After(entry.resetAt) {
			delete(l.entries, key)
		}
	}
	if len(l.entries) < maxMemoryKeys {
		return
	}
	// Still saturated after pruning: drop everything rather than grow forever.
	l.entries = make(map[string]*rateEntry)
}

func (l *RateLimiter) redisKey(key string) string {
	digest := sha256.Sum256([]byte(key))
	return "gopress:rl:" + l.name + ":" + hex.EncodeToString(digest[:12])
}
