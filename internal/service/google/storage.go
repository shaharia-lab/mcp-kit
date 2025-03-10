package google

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"os"
	"sync"
)

// GoogleOAuthTokenSourceStorage defines the interface for OAuth token storage
type GoogleOAuthTokenSourceStorage interface {
	// GetTokenSource returns an oauth2.TokenSource
	GetTokenSource(ctx context.Context) (oauth2.TokenSource, error)
	// SetTokenSource stores the token source for later use
	SetTokenSource(ctx context.Context, token *oauth2.Token, config *oauth2.Config) error
}

// InMemoryStorage implements GoogleOAuthTokenSourceStorage interface with in-memory storage
type InMemoryStorage struct {
	mu          sync.RWMutex
	token       *oauth2.Token
	oauthConfig *oauth2.Config
}

// NewInMemoryStorage creates a new instance of InMemoryStorage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

// GetTokenSource returns the token source from the in-memory storage
func (s *InMemoryStorage) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.token == nil || s.oauthConfig == nil {
		return nil, ErrNoTokenAvailable
	}

	// Create a token source that will automatically handle token refresh
	return s.oauthConfig.TokenSource(ctx, s.token), nil
}

// SetTokenSource stores the token source in the in-memory storage
func (s *InMemoryStorage) SetTokenSource(ctx context.Context, token *oauth2.Token, config *oauth2.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.token = token
	s.oauthConfig = config
	return nil
}

// FileTokenStorage implements GoogleOAuthTokenSourceStorage using a JSON file
type FileTokenStorage struct {
	filepath string
	mu       sync.RWMutex
}

// TokenData represents the structure we'll store in JSON
type TokenData struct {
	Token      *oauth2.Token `json:"token"`
	ConfigJSON []byte        `json:"config_json"`
}

// NewFileTokenStorage creates a new FileTokenStorage instance
func NewFileTokenStorage(filepath string) *FileTokenStorage {
	return &FileTokenStorage{
		filepath: filepath,
	}
}

func (s *FileTokenStorage) SetTokenSource(ctx context.Context, token *oauth2.Token, config *oauth2.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	data := TokenData{
		Token:      token,
		ConfigJSON: configJSON,
	}

	fileData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filepath, fileData, 0600)
}

func (s *FileTokenStorage) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return nil, err
	}

	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return nil, err
	}

	var config oauth2.Config
	if err := json.Unmarshal(tokenData.ConfigJSON, &config); err != nil {
		return nil, err
	}

	return config.TokenSource(ctx, tokenData.Token), nil
}
