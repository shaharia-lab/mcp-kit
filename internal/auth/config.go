package auth

import (
	"errors"
	"time"
)

// Config holds all authentication related configuration
type Config struct {
	AuthDomain       string        `envconfig:"AUTH_DOMAIN"`
	AuthClientID     string        `envconfig:"AUTH_CLIENT_ID"`
	AuthClientSecret string        `envconfig:"AUTH_CLIENT_SECRET"`
	AuthCallbackURL  string        `envconfig:"AUTH_CALLBACK_URL"`
	AuthTokenTTL     time.Duration `envconfig:"AUTH_TOKEN_TTL"`
	AuthAudience     string        `envconfig:"AUTH_AUDIENCE"`
}

// Validate ensures all required fields are present
func (c *Config) Validate() error {
	if c.AuthDomain == "" {
		return errors.New("auth domain is required")
	}
	if c.AuthClientID == "" {
		return errors.New("client ID is required")
	}
	if c.AuthClientSecret == "" {
		return errors.New("client secret is required")
	}
	if c.AuthCallbackURL == "" {
		return errors.New("callback URL is required")
	}
	if c.AuthAudience == "" {
		return errors.New("audience is required")
	}
	if c.AuthTokenTTL == 0 {
		c.AuthTokenTTL = 5 * time.Minute // Default TTL
	}
	return nil
}
