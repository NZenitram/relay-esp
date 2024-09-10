// middleware/auth.go
package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/nzenitram/relay-esp/models"
	"github.com/nzenitram/relay-esp/utils"
)

func APIKeyAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			apiKey := parts[1]
			user, err := models.GetUserByAPIKey(db, apiKey)
			if err != nil {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// Add the authenticated user to the request context
			ctx := context.WithValue(r.Context(), "authUser", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JWTAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			claims, err := utils.ValidateToken(bearerToken[1])
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			user, err := models.GetUserByID(db, claims.UserID)
			if err != nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
