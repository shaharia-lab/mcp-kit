package auth

import (
	"context"

	"golang.org/x/oauth2"
)

// AuthProvider defines the main interface for authentication operations
type AuthProvider interface {
	TokenValidator
	OAuth2Provider
}

// TokenValidator handles token validation
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) error
}

// OAuth2Provider handles OAuth2 operations
type OAuth2Provider interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
}
