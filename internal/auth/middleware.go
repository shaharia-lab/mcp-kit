package authenticator

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	"log"
	"net/http"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope string `json:"scope"`
}

// Validate satisfies the validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken is a middleware that validates JWTs in incoming requests.
func (a *Authenticator) EnsureValidToken() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the token from the Authorization header
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Remove 'Bearer ' prefix if present
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			// Verify the token
			ctx := r.Context()
			verifier := a.Verifier(&oidc.Config{
				ClientID: a.Config.ClientID,
			})

			idToken, err := verifier.Verify(ctx, token)
			if err != nil {
				log.Printf("Failed to verify token: %v", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add the verified token to the request context
			ctx = context.WithValue(ctx, "token", idToken)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
