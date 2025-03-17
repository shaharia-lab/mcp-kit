package cmd

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/mcp-kit/internal/prompt"
	"github.com/shaharia-lab/mcp-kit/internal/tools"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func NewServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		Long:  `Start the server with specified configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			ctx := context.Background()

			// Initialize all dependencies using Wire
			container, cleanup, err := InitializeAPI(ctx, configFile)
			if err != nil {
				return fmt.Errorf("failed to initialize application: %w", err)
			}
			defer cleanup()

			logger := container.LogrusLoggerImpl

			// Initialize the tracing service
			if err = container.TracingService.Initialize(ctx); err != nil {
				return fmt.Errorf("failed to initialize tracer: %w", err)
			}

			defer func() {
				if err = container.TracingService.Shutdown(ctx); err != nil {
					container.Logger.Printf("Error shutting down tracer: %v", err)
				}
			}()

			tracer := otel.Tracer("mcp-server")
			ctx, span := tracer.Start(ctx, "mcp_server_init")
			defer span.End()

			defer func() {
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
				}
			}()

			if err != nil {
				return fmt.Errorf("failed to create base server: %w", err)
			}

			googleAuthTokenSource, err := container.GoogleOAuthTokenSourceStorage.GetTokenSource(ctx)
			if err != nil {
				return fmt.Errorf("failed to get token source: %w", err)
			}

			// configure gmail service
			gmailSvc, err := gmail.NewService(ctx,
				option.WithTokenSource(googleAuthTokenSource),
				option.WithScopes(gmail.GmailReadonlyScope),
			)
			if err != nil {
				logger.Fatalf("Failed to create Gmail service: %v", err)
			}

			err = container.BaseMCPServer.AddPrompts(prompt.MCPPromptsRegistry...)
			if err != nil {
				return fmt.Errorf("failed to add prompts: %w", err)
			}

			toolsRegistry := tools.NewRegistry(container.Config.Tools, container.LogrusLoggerImpl, gmailSvc)
			err = toolsRegistry.Init()
			if err != nil {
				return fmt.Errorf("failed to initialize tools registry: %w", err)
			}

			err = container.BaseMCPServer.AddTools(toolsRegistry.GetToolLists()...)
			if err != nil {
				return fmt.Errorf("failed to add tools: %w", err)
			}

			server := mcp.NewSSEServer(container.BaseMCPServer)
			server.SetAddress(fmt.Sprintf(":%d", container.Config.MCPServerPort))

			container.LogrusLoggerImpl.Info("Server is starting...")
			if err = server.Run(ctx); err != nil {
				return fmt.Errorf("failed to run server: %w", err)
			}

			return nil
		},
	}
}
