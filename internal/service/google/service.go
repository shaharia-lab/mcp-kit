package google

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/shaharia-lab/mcp-kit/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleService handles OAuth2 flow for Google services
type GoogleService struct {
	storage     GoogleOAuthTokenSourceStorage
	oauthConfig *oauth2.Config
	stateCookie string
	redirectURL string
	enabled     bool
}

// NewGoogleService creates a new instance of GoogleService
func NewGoogleService(storage GoogleOAuthTokenSourceStorage, config config.GoogleConfig) *GoogleService {
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint:     google.Endpoint,
	}

	return &GoogleService{
		storage:     storage,
		oauthConfig: oauthConfig,
		stateCookie: config.StateCookie,
		redirectURL: config.RedirectURL,
		enabled:     config.Enabled,
	}
}

// HandleOAuthStart initiates the OAuth2 flow
func (s *GoogleService) HandleOAuthStart(w http.ResponseWriter, r *http.Request) {
	if !s.enabled {
		http.Error(w, "Google services are not enabled", http.StatusForbidden)
		return
	}

	state, err := generateRandomState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Set state in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     s.stateCookie,
		Value:    state,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	})

	// Redirect to Google's consent page
	url := s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleOAuthCallback handles the OAuth2 callback and stores the token
func (s *GoogleService) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if !s.enabled {
		http.Error(w, "Google services are not enabled", http.StatusForbidden)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, ErrInvalidState.Error(), http.StatusBadRequest)
		return
	}

	// Verify state from cookie
	cookie, err := r.Cookie(s.stateCookie)
	if err != nil || cookie.Value != state {
		http.Error(w, ErrInvalidState.Error(), http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, ErrMissingCode.Error(), http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := s.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Store the token
	if err := s.storage.SetTokenSource(r.Context(), token, s.oauthConfig); err != nil {
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     s.stateCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
	})

	// Redirect to success page or show success message
	w.Write([]byte("Authentication successful"))
}

// generateRandomState generates a random state string for OAuth security
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetTokenSource returns the current token source from storage
func (s *GoogleService) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	return s.storage.GetTokenSource(ctx)
}
