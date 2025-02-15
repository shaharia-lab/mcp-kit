package cmd

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/mcp-kit/pkg/config"
	"github.com/spf13/cobra"
	"log"
	"time"
)

func NewClientCmd(ctx context.Context, logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "client",
		Short: "Start the client",
		Long:  `Start the client and connect to the server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			sseConfig := mcp.ClientConfig{
				ClientName:    "My MCP Kit Client",
				ClientVersion: "1.0.0",
				Logger:        log.New(logger.Writer(), "", log.LstdFlags),
				RetryDelay:    5 * time.Second,
				MaxRetries:    3,
				SSE: mcp.SSEConfig{
					URL: cfg.MCPServerURL,
				},
			}

			sseTransport := mcp.NewSSETransport()
			sseClient := mcp.NewClient(sseTransport, sseConfig)

			if err := sseClient.Connect(); err != nil {
				log.Fatalf("SSE Client failed to connect: %v", err)
			}
			defer sseClient.Close()

			tools, err := sseClient.ListTools()
			if err != nil {
				return fmt.Errorf("failed to list mcpToolsRegistry (SSE): %w", err)
			}
			fmt.Printf("SSE Tools: %+v\n", tools)

			return nil
		},
	}
}
