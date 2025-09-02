package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID      int      `json:"user_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

type contextKey string

const userDetailsKey contextKey = "userDetails"

func GenerateToken(userID int, role string, permissions []string) (string, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return "", fmt.Errorf("JWT_SECRET_KEY environment variable not set")
	}

	claims := UserClaims{
		userID,
		role,
		permissions,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func ValidateToken(jwtString string) (*UserClaims, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("JWT_SECRET_KEY environment variable not set")
	}

	token, err := jwt.ParseWithClaims(jwtString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid or expired token")
}

func AuthMiddleware(nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // No "Bearer " prefix found
			respondWithError(w, "Invalid token format (expected 'Bearer <token>')", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			respondWithError(w, "Invalid or expired token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userDetailsKey, claims)
		nextHandler.ServeHTTP(w, r.WithContext(ctx))
	}
}
func RequirePermission(permission string, nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the claims from the request context
		userDetails, ok := GetUserDetailsFromContext(r.Context())
		if !ok {
			respondWithError(w, "Authentication context missing", http.StatusForbidden)
			return
		}

		// Check if the required permission exists in the user's permissions slice
		hasPermission := false
		for _, p := range userDetails.Permissions {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			log.Printf("User lacks required permission: %s", permission)
			respondWithError(w, "Forbidden", http.StatusForbidden)
			return
		}

		// If the user has the permission, call the next handler
		nextHandler.ServeHTTP(w, r)
	}
}

func GetUserDetailsFromContext(ctx context.Context) (userDetails, bool) {
	userDetails, ok := ctx.Value(userIDContextKey).(*UserClaims)
	return userDetails, ok
}
