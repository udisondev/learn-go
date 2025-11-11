package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements in-memory rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*client
	limit    int           // max requests
	window   time.Duration // time window
	cleanupInterval time.Duration
}

type client struct {
	requests []time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// limit: maximum number of requests
// window: time window (e.g., 15 minutes)
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		limit:   limit,
		window:  window,
		cleanupInterval: 5 * time.Minute,
	}

	// Start cleanup goroutine to remove old entries
	go rl.cleanup()

	return rl
}

// Middleware returns HTTP middleware that limits requests by IP
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		if !rl.allow(ip) {
			http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AllowEmail checks if email-based rate limit allows the request
func (rl *RateLimiter) AllowEmail(email string) bool {
	key := "email:" + email
	return rl.allow(key)
}

// allow checks if the request is allowed
func (rl *RateLimiter) allow(key string) bool {
	now := time.Now()

	rl.mu.Lock()
	c, exists := rl.clients[key]
	if !exists {
		c = &client{
			requests: []time.Time{},
		}
		rl.clients[key] = c
	}
	rl.mu.Unlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove requests outside the time window
	cutoff := now.Add(-rl.window)
	validRequests := []time.Time{}
	for _, t := range c.requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	c.requests = validRequests

	// Check if limit exceeded
	if len(c.requests) >= rl.limit {
		return false
	}

	// Add current request
	c.requests = append(c.requests, now)
	return true
}

// cleanup periodically removes old clients
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)

		for key, c := range rl.clients {
			c.mu.Lock()
			// Remove if no requests in the window
			if len(c.requests) == 0 || c.requests[len(c.requests)-1].Before(cutoff) {
				delete(rl.clients, key)
			}
			c.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// getIP extracts IP address from request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
