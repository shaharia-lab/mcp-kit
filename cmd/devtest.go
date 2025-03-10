package cmd

import (
	"context"
	"github.com/spf13/cobra"
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

			container, _, err := InitializeAPI(ctx, "")
			if err != nil {
				return err
			}

			// Set up routes
			http.HandleFunc("/oauth/login", container.GoogleService.HandleOAuthStart)
			http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
				container.GoogleService.HandleOAuthCallback(w, r)
				w.Write([]byte("Authentication successful and Gmail API test completed"))
			})

			log.Printf("Server starting on :8080. Visit http://localhost:9090/oauth/login to start OAuth flow")
			return http.ListenAndServe(":9090", nil)
		},
	}
}
