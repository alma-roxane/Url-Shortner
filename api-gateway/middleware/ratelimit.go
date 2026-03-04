package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count      int
	windowFrom time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	limit    int
	visitors map[string]visitor
}

func NewRateLimiter(limit int) *RateLimiter {
	if limit <= 0 {
		limit = 120
	}
	return &RateLimiter{limit: limit, visitors: make(map[string]visitor)}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		now := time.Now().UTC()

		rl.mu.Lock()
		v := rl.visitors[ip]
		if now.Sub(v.windowFrom) > time.Minute {
			v = visitor{windowFrom: now, count: 0}
		}
		v.count++
		rl.visitors[ip] = v
		rl.mu.Unlock()

		if v.count > rl.limit {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
