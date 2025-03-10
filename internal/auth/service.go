package auth

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/mcp-kit/internal/config"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/shaharia-lab/goai/observability"
	"golang.org/x/oauth2"
)

// AuthService implements AuthProvider interface
type AuthService struct {
	provider    *oidc.Provider
	config      config.AuthConfig
	oauthConfig oauth2.Config
	logger      observability.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(ctx context.Context, cfg config.AuthConfig, logger observability.Logger) (*AuthService, error) {
	provider, err := oidc.NewProvider(ctx, "https://"+cfg.AuthDomain+"/")
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauthConfig := oauth2.Config{
		ClientID:     cfg.AuthClientID,
		ClientSecret: cfg.AuthClientSecret,
		RedirectURL:  cfg.AuthCallbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &AuthService{
		provider:    provider,
		config:      cfg,
		oauthConfig: oauthConfig,
		logger:      logger,
	}, nil
}

// ValidateToken implements TokenValidator interface
func (a *AuthService) ValidateToken(ctx context.Context, tokenString string) error {
	ctx, cancel := context.WithTimeout(ctx, a.config.AuthTokenTTL)
	defer cancel()

	issuerURL, err := url.Parse("https://" + a.config.AuthDomain + "/")
	if err != nil {
		return fmt.Errorf("failed to parse issuer URL: %w", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{a.config.AuthAudience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to set up JWT validator: %w", err)
	}

	_, err = jwtValidator.ValidateToken(ctx, tokenString)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	return nil
}

// AuthCodeURL implements OAuth2Provider interface
func (a *AuthService) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return a.oauthConfig.AuthCodeURL(state, opts...)
}

// Exchange implements OAuth2Provider interface
func (a *AuthService) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, a.config.AuthTokenTTL)
	defer cancel()

	return a.oauthConfig.Exchange(ctx, code)
}

// CustomClaims contains custom data we want from the token
type CustomClaims struct {
	Scope string `json:"scope"`
}

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func (c CustomClaims) HasScope(expectedScope string) bool {
	for _, scope := range strings.Split(c.Scope, " ") {
		if scope == expectedScope {
			return true
		}
	}
	return false
}
