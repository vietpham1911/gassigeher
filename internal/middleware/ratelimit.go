package middleware

import (
	"net/http"
	"sync"
	"time"
)

// BUG FIX #6: Simple rate limiting for login endpoint
// Prevents brute force attacks

type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

var loginLimiter = &rateLimiter{
	requests: make(map[string][]time.Time),
	limit:    5,                 // 5 attempts
	window:   1 * time.Minute,   // per minute
}

// RateLimitLogin limits login attempts per IP address
func RateLimitLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		loginLimiter.mu.Lock()
		defer loginLimiter.mu.Unlock()

		now := time.Now()

		// Clean old requests outside window
		if requests, exists := loginLimiter.requests[ip]; exists {
			validRequests := []time.Time{}
			for _, reqTime := range requests {
				if now.Sub(reqTime) < loginLimiter.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			loginLimiter.requests[ip] = validRequests
		}

		// Check if limit exceeded
		if len(loginLimiter.requests[ip]) >= loginLimiter.limit {
			http.Error(w, `{"error":"Zu viele Anmeldeversuche. Bitte versuchen Sie es in einer Minute erneut."}`, http.StatusTooManyRequests)
			return
		}

		// Add current request
		loginLimiter.requests[ip] = append(loginLimiter.requests[ip], now)

		next.ServeHTTP(w, r)
	})
}

// DONE: BUG #6 FIXED - Rate limiting implemented for login endpoint
