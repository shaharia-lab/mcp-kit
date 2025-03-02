package cmd

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/goai/mcp"
	goaiObs "github.com/shaharia-lab/goai/observability"
	"github.com/shaharia-lab/mcp-kit/internal/config"
	"github.com/shaharia-lab/mcp-kit/internal/prompt"
	"github.com/shaharia-lab/mcp-kit/internal/tools"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"log"
)

func NewServerCmd(logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		Long:  `Start the server with specified configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			ctx := context.Background()

			// Initialize all dependencies using Wire
			container, cleanup, err := InitializeAPI(ctx)
			if err != nil {
				return fmt.Errorf("failed to initialize application: %w", err)
			}
			defer cleanup()

			// Initialize the tracing service
			tracingService := ProvideTracingService(container.Config, logrus.New())
			if err = tracingService.Initialize(ctx); err != nil {
				return fmt.Errorf("failed to initialize tracer: %w", err)
			}

			defer func() {
				if err = tracingService.Shutdown(ctx); err != nil {
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

			l := goaiObs.NewDefaultLogger()
			baseServer, err := mcp.NewBaseServer(
				mcp.UseLogger(l),
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

			l.Info("Server is starting...")
			if err = server.Run(ctx); err != nil {
				return fmt.Errorf("failed to run server: %w", err)
			}

			return nil
		},
	}
}
