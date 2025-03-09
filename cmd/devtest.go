package cmd

import (
	"context"
	"github.com/shaharia-lab/mcp-kit/internal/service/google"
	"github.com/spf13/cobra"
	"google.golang.org/api/gmail/v1"
	"log"
	"net/http"
)

func NewDevTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "devtest",
		Short: "Run the development test",
		Long:  `Run the development test`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			container, _, err := InitializeAPI(ctx)
			if err != nil {
				return err
			}

			googleOAuthStorage := google.NewFileTokenStorage(container.Config.GoogleServiceConfig.TokenSourceFile)

			// Configure Google Service
			googleService := google.NewGoogleService(googleOAuthStorage, google.Config{
				ClientID:     container.Config.GoogleServiceConfig.ClientID,
				ClientSecret: container.Config.GoogleServiceConfig.ClientSecret,
				RedirectURL:  "http://localhost:9090/oauth/callback",
				Scopes: []string{
					gmail.GmailReadonlyScope,
					gmail.GmailSendScope,
					gmail.GmailModifyScope,
					"openid",
					"https://www.googleapis.com/auth/userinfo.email",
					"https://www.googleapis.com/auth/userinfo.profile",
				},

				StateCookie: "google-oauth-state",
			})

			// Set up routes
			http.HandleFunc("/oauth/login", googleService.HandleOAuthStart)
			http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
				googleService.HandleOAuthCallback(w, r)
				w.Write([]byte("Authentication successful and Gmail API test completed"))
			})

			log.Printf("Server starting on :8080. Visit http://localhost:9090/oauth/login to start OAuth flow")
			return http.ListenAndServe(":9090", nil)
		},
	}
}
