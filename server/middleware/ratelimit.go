package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Limiter is a fixed-window, in-memory rate limiter keyed by an arbitrary
// string. It is per-process: run a shared store (e.g. Redis) if you scale out.
type Limiter struct {
	mu        sync.Mutex
	max       int
	window    time.Duration
	hits      map[string]*bucket
	lastSweep time.Time
}

type bucket struct {
	count int
	reset time.Time
}

func NewLimiter(max int, window time.Duration) *Limiter {
	return &Limiter{max: max, window: window, hits: map[string]*bucket{}, lastSweep: time.Now()}
}

// Allow records a hit for key and reports whether it is within the limit.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if now.Sub(l.lastSweep) > l.window {
		for k, b := range l.hits {
			if now.After(b.reset) {
				delete(l.hits, k)
			}
		}
		l.lastSweep = now
	}

	b, ok := l.hits[key]
	if !ok || now.After(b.reset) {
		b = &bucket{reset: now.Add(l.window)}
		l.hits[key] = b
	}
	b.count++
	return b.count <= l.max
}

// Reset forgets the hits recorded for key (e.g. after a successful login).
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	delete(l.hits, key)
	l.mu.Unlock()
}

// PerIP rejects requests with 429 once a client IP exceeds the limit.
func (l *Limiter) PerIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.Allow(c.ClientIP()) {
			TooManyRequests(c)
			return
		}
		c.Next()
	}
}

func TooManyRequests(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts, try again later"})
}
