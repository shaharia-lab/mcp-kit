package cmd

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/mcp-kit/pkg/config"
	"github.com/shaharia-lab/mcp-kit/pkg/prompt"
	"github.com/shaharia-lab/mcp-kit/pkg/tools"
	"github.com/spf13/cobra"
	"log"
)

func NewServerCmd(logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		Long:  `Start the server with specified configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			baseServer, err := mcp.NewBaseServer(
				mcp.UseLogger(log.New(logger.Writer(), "", log.LstdFlags)),
			)

			if err != nil {
				return fmt.Errorf("failed to create base server: %w", err)
			}

			err = baseServer.AddPrompts(prompt.MCPPromptsRegistry...)
			if err != nil {
				return fmt.Errorf("failed to add prompts: %w", err)
			}

			err = baseServer.AddTools(tools.MCPToolsRegistry...)
			if err != nil {
				return fmt.Errorf("failed to add tools: %w", err)
			}

			server := mcp.NewSSEServer(baseServer)
			server.SetAddress(fmt.Sprintf(":%d", cfg.MCPServerPort))

			if err := server.Run(ctx); err != nil {
				return fmt.Errorf("failed to run server: %w", err)
			}

			return nil
		},
	}
}
