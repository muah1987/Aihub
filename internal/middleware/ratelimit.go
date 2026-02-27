package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type visitor struct {
	tokens    float64
	lastSeen  time.Time
}

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     float64
	burst    float64
}

func NewRateLimiter(ratePerSecond, burst float64) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     ratePerSecond,
		burst:    burst,
	}

	// Cleanup stale entries every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

// clientIP extracts the real client IP from the request.
// It trusts X-Real-IP and X-Forwarded-For headers, so this function should
// only be used when the application is deployed behind a reverse proxy (e.g.
// Nginx) that strips or overwrites these headers from untrusted clients before
// forwarding requests. Without such a proxy, clients can spoof their IP and
// bypass rate limiting.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		if parsed := net.ParseIP(strings.TrimSpace(ip)); parsed != nil {
			return parsed.String()
		}
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		first := strings.TrimSpace(strings.SplitN(forwarded, ",", 2)[0])
		if parsed := net.ParseIP(first); parsed != nil {
			return parsed.String()
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			v = &visitor{tokens: rl.burst}
			rl.visitors[ip] = v
		}

		elapsed := time.Since(v.lastSeen).Seconds()
		v.tokens += elapsed * rl.rate
		if v.tokens > rl.burst {
			v.tokens = rl.burst
		}
		v.lastSeen = time.Now()

		if v.tokens < 1 {
			rl.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}

		v.tokens--
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
