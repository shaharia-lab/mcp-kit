package cmd

import (
	"context"
	"fmt"
	mcptools "github.com/shaharia-lab/mcp-tools"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"log"

	"github.com/shaharia-lab/goai/mcp"
	goaiObs "github.com/shaharia-lab/goai/observability"
	"github.com/shaharia-lab/mcp-kit/internal/config"
	"github.com/shaharia-lab/mcp-kit/internal/prompt"
	"github.com/shaharia-lab/mcp-kit/internal/tools"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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
			toolsLists := setupTools(container.LogrusLoggerImpl, gmailSvc)

			err = container.BaseMCPServer.AddPrompts(prompt.MCPPromptsRegistry...)
			if err != nil {
				return fmt.Errorf("failed to add prompts: %w", err)
			}

			err = container.BaseMCPServer.AddTools(toolsLists...)
			if err != nil {
				return fmt.Errorf("failed to add tools: %w", err)
			}

			server := mcp.NewSSEServer(container.BaseMCPServer)
			server.SetAddress(fmt.Sprintf(":%d", cfg.MCPServerPort))

			container.LogrusLoggerImpl.Info("Server is starting...")
			if err = server.Run(ctx); err != nil {
				return fmt.Errorf("failed to run server: %w", err)
			}

			return nil
		},
	}
}

func setupTools(logger goaiObs.Logger, gmailService *gmail.Service) []mcp.Tool {
	ts := tools.MCPToolsRegistry

	ghConfig := mcptools.NewGitHubTool(logger, mcptools.GitHubConfig{})
	fileSystem := mcptools.NewFileSystem(logger, mcptools.FileSystemConfig{})
	docker := mcptools.NewDocker(logger)
	git := mcptools.NewGit(logger, mcptools.GitConfig{})
	curl := mcptools.NewCurl(logger, mcptools.CurlConfig{})
	postgres := mcptools.NewPostgreSQL(logger, mcptools.PostgreSQLConfig{})

	gm := mcptools.NewGmail(logger, gmailService, mcptools.GmailConfig{})

	ts = append(
		ts,

		// Curl tools
		curl.CurlAllInOneTool(),

		// Git tools
		git.GitAllInOneTool(),

		// Docker tools
		docker.DockerAllInOneTool(),

		// File system tools
		fileSystem.FileSystemAllInOneTool(),

		// GitHub tools
		ghConfig.GetIssuesTool(),
		ghConfig.GetPullRequestsTool(),
		ghConfig.GetRepositoryTool(),
		ghConfig.GetSearchTool(),

		// PostgreSQL tools
		postgres.PostgreSQLAllInOneTool(),

		// Gmail
		gm.GmailAllInOneTool(),
	)

	return ts
}
