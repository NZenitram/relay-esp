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

type contextKey string

const (
	AuthUserKey contextKey = "authUser"
)

// func APIKeyAuth(db *sql.DB) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			authHeader := r.Header.Get("Authorization")
// 			log.Printf("Auth Header: %s", authHeader)

// 			if authHeader == "" {
// 				log.Println("Missing Authorization header")
// 				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
// 				return
// 			}

// 			parts := strings.Split(authHeader, " ")
// 			if len(parts) != 2 || parts[0] != "Bearer" {
// 				log.Printf("Invalid header format. Parts: %v", parts)
// 				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
// 				return
// 			}

// 			apiKey := parts[1]
// 			log.Printf("Attempting to get user with API key: %s", apiKey)

// 			user, err := models.GetUserByAPIKey(db, apiKey)
// 			if err != nil {
// 				log.Printf("Error getting user by API key: %v", err)
// 				http.Error(w, "Invalid API key", http.StatusUnauthorized)
// 				return
// 			}

// 			log.Printf("User authenticated: %s", user.Username)

// 			ctx := context.WithValue(r.Context(), AuthUserKey, user)
// 			next.ServeHTTP(w, r.WithContext(ctx))
// 		})
// 	}
// }

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

			ctx := context.WithValue(r.Context(), AuthUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
