// middleware/auth.go
package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/nzenitram/relay-esp/models"
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
