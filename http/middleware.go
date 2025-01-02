package http

import (
	"context"
	"log"
	"net/http"
	"time"
)

// loggingMiddleware is a middleware that logs the request method, path, and duration
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
		}()
		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware is a middleware that implements rate limiting logic
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement rate limiting logic here
		next.ServeHTTP(w, r)
	})
}

// authMiddleware is a middleware that checks for an authorization header
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for auth token in header
		token := r.Header.Get("Authorization")
		if token == "" {
			respondWithError(w, http.StatusUnauthorized, "Missing authorization token")
			return
		}

		// TODO: Implement token validation

		// add token to context
		ctx := context.WithValue(r.Context(), "token", token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
