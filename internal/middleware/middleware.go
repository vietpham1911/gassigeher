package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tranm/gassigeher/internal/services"
)

type contextKey string

const UserIDKey contextKey = "userID"
const EmailKey contextKey = "email"
const IsAdminKey contextKey = "isAdmin"

// LoggingMiddleware logs HTTP requests
// BUG FIX #13: Sanitize sensitive data from logs
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Sanitize URL for logging (don't log tokens in query params)
		sanitizedPath := r.URL.Path
		if r.URL.RawQuery != "" {
			// Redact sensitive query parameters
			if strings.Contains(r.URL.RawQuery, "token") {
				sanitizedPath += "?token=REDACTED"
			} else {
				sanitizedPath += "?" + r.URL.RawQuery
			}
		}

		log.Printf("%s %s", r.Method, sanitizedPath)
		next.ServeHTTP(w, r)
		log.Printf("Completed in %v", time.Since(start))
	})
}
// DONE: BUG #13 FIXED - Sensitive data redacted from logs

// CORSMiddleware adds CORS headers
// BUG FIX #1: Restrict CORS to specific origins instead of "*"
// Accepts baseURL from config for dynamic CORS origin configuration
func CORSMiddleware(baseURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Default to localhost if baseURL not provided
			if baseURL == "" {
				baseURL = "http://localhost:8080"
			}

			// Allowed origins for CORS (configurable base + additional domains)
			allowedOrigins := []string{
				baseURL,
				"https://gassi.cuong.net",
				"https://www.gassi.cuong.net",
			}

			origin := r.Header.Get("Origin")
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}

			// If no origin header or not in allowed list, allow same-origin requests
			if origin == "" {
				w.Header().Set("Access-Control-Allow-Origin", baseURL)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// DONE: BUG #1 FIXED - CORS now restricted to specific allowed origins

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"Missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"Invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			authService := services.NewAuthService(jwtSecret, 24) // expiration not used here
			claims, err := authService.ValidateJWT(tokenString)
			if err != nil {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized) // BUG FIX #3
				return
			}

			// Extract claims
			userID, ok := (*claims)["user_id"].(float64)
			if !ok {
				http.Error(w, `{"error":"Invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			email, ok := (*claims)["email"].(string)
			if !ok {
				http.Error(w, `{"error":"Invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			isAdmin, ok := (*claims)["is_admin"].(bool)
			if !ok {
				isAdmin = false
			}

			// Add to context
			ctx := context.WithValue(r.Context(), UserIDKey, int(userID))
			ctx = context.WithValue(ctx, EmailKey, email)
			ctx = context.WithValue(ctx, IsAdminKey, isAdmin)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin middleware checks if user is an admin
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, ok := r.Context().Value(IsAdminKey).(bool)
		if !ok || !isAdmin {
			http.Error(w, `{"error":"Admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Enforce HTTPS in production
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")

		next.ServeHTTP(w, r)
	})
}
