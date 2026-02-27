package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const (
	ContextKeyUserID contextKey = "userID"
	ContextKeyEmail  contextKey = "email"
	ContextKeyRole   contextKey = "role"
	ContextKeyJWT    contextKey = "jwtService"
)

// InjectJWTService adds the JWT service to the request context for WebSocket handlers.
func InjectJWTService(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ContextKeyJWT, jwtService)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Middleware(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authorization header required"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization format"})
				return
			}

			claims, err := jwtService.ValidateToken(parts[1])
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}

			if claims.Type != AccessToken {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token type"})
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extracts the user ID from the request context.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return userID, ok
}

// GetJWTServiceFromContext retrieves the injected JWTService from context.
func GetJWTServiceFromContext(ctx context.Context) *JWTService {
	svc, _ := ctx.Value(ContextKeyJWT).(*JWTService)
	return svc
}

// OptionalAuth extracts auth info if present but doesn't require it.
func OptionalAuth(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				claims, err := jwtService.ValidateToken(parts[1])
				if err == nil && claims.Type == AccessToken {
					ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
					ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
					ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
