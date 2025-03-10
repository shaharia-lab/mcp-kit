package cmd

import (
	"context"
	"github.com/shaharia-lab/mcp-kit/internal/service/google"
	"log"
	"time"

	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	goaiObs "github.com/shaharia-lab/goai/observability"
	"github.com/shaharia-lab/mcp-kit/internal/auth"
	"github.com/shaharia-lab/mcp-kit/internal/config"
	"github.com/shaharia-lab/mcp-kit/internal/observability"
	"github.com/sirupsen/logrus"
)

// Container holds all the dependencies for our application
type Container struct {
	Logger                        *log.Logger
	MCPClient                     *mcp.Client
	ToolsProvider                 *goai.ToolsProvider
	ChatHistoryStorage            goai.ChatHistoryStorage
	Config                        *config.Config
	TracingService                *observability.TracingService
	LoggerLogrus                  *logrus.Logger
	LogrusLoggerImpl              goaiObs.Logger
	BaseMCPServer                 *mcp.BaseServer
	AuthService                   *auth.AuthService
	AuthMiddleware                *auth.AuthMiddleware
	GoogleService                 *google.GoogleService
	GoogleOAuthTokenSourceStorage google.GoogleOAuthTokenSourceStorage
}

func ProvideLogger() *log.Logger {
	return log.Default()
}

func ProvideConfig(configFilePath string) (*config.Config, error) {
	return loadConfig(configFilePath)
}

func ProvideMCPClient(l goaiObs.Logger, cfg *config.Config) *mcp.Client {
	return mcp.NewClient(mcp.NewSSETransport(l), mcp.ClientConfig{
		ClientName:          "My MCP Kit Client",
		ClientVersion:       "1.0.0",
		Logger:              l,
		RetryDelay:          3 * time.Second,
		MaxRetries:          5,
		HealthCheckInterval: 15 * time.Second,
		ConnectionTimeout:   60 * time.Second,
		SSE: mcp.SSEConfig{
			URL: cfg.MCPServerURL,
		},
		RequestTimeout: 120 * time.Second,
	})
}

func ProvideLogrusLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		// Customize timestamp format
		TimestampFormat: time.RFC3339,
		// Include caller information
		PrettyPrint: false,
		// Disable HTML escape of special characters
		DisableHTMLEscape: true,
		// Include full timestamp
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	return logger
}

func ProvideLogrusLoggerImpl(logger *logrus.Logger) goaiObs.Logger {
	return observability.NewLogrusLogger(logger)
}

func provideAuthenticator(ctx context.Context, cfg *config.Config, logger goaiObs.Logger) (*auth.AuthService, error) {
	return auth.NewAuthService(ctx, cfg.Auth, logger)
}

func ProvideToolsProvider(mcpClient *mcp.Client) (*goai.ToolsProvider, error) {
	provider := goai.NewToolsProvider()
	if err := provider.AddMCPClient(mcpClient); err != nil {
		return nil, err
	}
	return provider, nil
}

func ProvideTracingService(cfg *config.Config, logger goaiObs.Logger) *observability.TracingService {
	tracingConfig := config.TracingConfig{
		Enabled:         cfg.Tracing.Enabled,
		ServiceName:     cfg.Tracing.ServiceName,
		EndpointAddress: cfg.Tracing.EndpointAddress,
		Timeout:         cfg.Tracing.Timeout,
		SamplingRate:    cfg.Tracing.SamplingRate,
		BatchTimeout:    cfg.Tracing.BatchTimeout,
		Environment:     cfg.Tracing.Environment,
		Version:         cfg.Tracing.Version,
	}

	return observability.NewTracingService(tracingConfig, logger)
}

func ProvideChatHistoryStorage() goai.ChatHistoryStorage {
	return goai.NewInMemoryChatHistoryStorage()
}

func ProvideMCPBaseServer(l goaiObs.Logger) (*mcp.BaseServer, error) {
	return mcp.NewBaseServer(
		mcp.UseLogger(l),
	)
}

func ProvideGoogleOAuthTokenSourceStorage(cfg *config.Config) google.GoogleOAuthTokenSourceStorage {
	return google.NewFileTokenStorage(cfg.GoogleServiceConfig.TokenSourceFile)
}

func ProvideGoogleService(cfg *config.Config, oauthTokenStorage google.GoogleOAuthTokenSourceStorage) *google.GoogleService {
	return google.NewGoogleService(oauthTokenStorage, cfg.GoogleServiceConfig)
}

func NewContainer(
	logger *log.Logger,
	mcpClient *mcp.Client,
	toolsProvider *goai.ToolsProvider,
	chatHistoryStorage goai.ChatHistoryStorage,
	config *config.Config,
	tracingService *observability.TracingService,
	loggerLogrus *logrus.Logger,
	logrusLoggerImpl goaiObs.Logger,
	baseServer *mcp.BaseServer,
	authService *auth.AuthService,
	googleService *google.GoogleService,
	googleOAuthTokenSourceStorage google.GoogleOAuthTokenSourceStorage,
) *Container {
	return &Container{
		Logger:                        logger,
		MCPClient:                     mcpClient,
		ToolsProvider:                 toolsProvider,
		ChatHistoryStorage:            chatHistoryStorage,
		Config:                        config,
		TracingService:                tracingService,
		LoggerLogrus:                  loggerLogrus,
		LogrusLoggerImpl:              logrusLoggerImpl,
		BaseMCPServer:                 baseServer,
		AuthService:                   authService,
		AuthMiddleware:                auth.NewAuthMiddleware(authService, logrusLoggerImpl),
		GoogleService:                 googleService,
		GoogleOAuthTokenSourceStorage: googleOAuthTokenSourceStorage,
	}
}
