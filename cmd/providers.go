package cmd

import (
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	goaiObs "github.com/shaharia-lab/goai/observability"
	"github.com/shaharia-lab/mcp-kit/internal/config"
	"github.com/shaharia-lab/mcp-kit/internal/observability"
	"github.com/sirupsen/logrus"
	"log"
	"time"
)

// Container holds all the dependencies for our application
type Container struct {
	Logger             *log.Logger
	MCPClient          *mcp.Client
	ToolsProvider      *goai.ToolsProvider
	ChatHistoryStorage goai.ChatHistoryStorage
	Config             *config.Config
	TracingService     *observability.TracingService
	LoggerLogrus       *logrus.Logger
	LogrusLoggerImpl   goaiObs.Logger
}

func ProvideLogger() *log.Logger {
	return log.Default()
}

func ProvideConfig() (*config.Config, error) {
	return loadConfig()
}

func ProvideMCPClient(cfg *config.Config) *mcp.Client {
	l := goaiObs.NewDefaultLogger()

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
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return logger
}

func ProvideLogrusLoggerImpl(logger *logrus.Logger) goaiObs.Logger {
	return observability.NewLogrusLogger(logger)
}

func ProvideToolsProvider(mcpClient *mcp.Client) (*goai.ToolsProvider, error) {
	provider := goai.NewToolsProvider()
	if err := provider.AddMCPClient(mcpClient); err != nil {
		return nil, err
	}
	return provider, nil
}

func ProvideTracingService(cfg *config.Config, logger *logrus.Logger) *observability.TracingService {
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

func NewContainer(
	logger *log.Logger,
	mcpClient *mcp.Client,
	toolsProvider *goai.ToolsProvider,
	chatHistoryStorage goai.ChatHistoryStorage,
	config *config.Config,
	tracingService *observability.TracingService,
	loggerLogrus *logrus.Logger,
	logrusLoggerImpl goaiObs.Logger,
) *Container {
	return &Container{
		Logger:             logger,
		MCPClient:          mcpClient,
		ToolsProvider:      toolsProvider,
		ChatHistoryStorage: chatHistoryStorage,
		Config:             config,
		TracingService:     tracingService,
		LoggerLogrus:       loggerLogrus,
		LogrusLoggerImpl:   logrusLoggerImpl,
	}
}
