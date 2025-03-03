package auth

import (
	"github.com/shaharia-lab/goai/observability"
	"net/http"
	"strings"
)

// AuthMiddleware handles authentication for HTTP requests
type AuthMiddleware struct {
	validator TokenValidator
	logger    observability.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(validator TokenValidator, logger observability.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		validator: validator,
		logger:    logger,
	}
}

// EnsureValidToken is a middleware that ensures a valid token is present
func (am *AuthMiddleware) EnsureValidToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			am.logger.Error("no token found", nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := am.validator.ValidateToken(r.Context(), token); err != nil {
			am.logger.Error("invalid token", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractToken extracts the token from the Authorization header
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}
