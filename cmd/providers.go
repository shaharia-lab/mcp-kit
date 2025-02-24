package cmd

import (
	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	goaiObs "github.com/shaharia-lab/goai/observability"
	"github.com/shaharia-lab/mcp-kit/config"
	"github.com/shaharia-lab/mcp-kit/storage"
	"log"
	"time"
)

// Container holds all the dependencies for our application
type Container struct {
	Logger             *log.Logger
	MCPClient          *mcp.Client
	ToolsProvider      *goai.ToolsProvider
	ChatHistoryStorage storage.ChatHistoryStorage
	Router             *chi.Mux
	Config             *config.Config
}

func ProvideLogger() *log.Logger {
	return log.Default()
}

func ProvideConfig() (*config.Config, error) {
	return loadConfig()
}

func ProvideMCPClient(logger *log.Logger, cfg *config.Config) *mcp.Client {
	l := goaiObs.NewDefaultLogger()
	return mcp.NewClient(mcp.NewSSETransport(l), mcp.ClientConfig{
		ClientName:    "My MCP Kit Client",
		ClientVersion: "1.0.0",
		Logger:        l,
		RetryDelay:    5 * time.Second,
		MaxRetries:    3,
		SSE: mcp.SSEConfig{
			URL: cfg.MCPServerURL,
		},
		RequestTimeout: 60 * time.Second,
	})
}

func ProvideToolsProvider(mcpClient *mcp.Client) (*goai.ToolsProvider, error) {
	provider := goai.NewToolsProvider()
	if err := provider.AddMCPClient(mcpClient); err != nil {
		return nil, err
	}
	return provider, nil
}

func ProvideChatHistoryStorage() storage.ChatHistoryStorage {
	return storage.NewInMemoryChatHistoryStorage()
}

func ProvideRouter(
	mcpClient *mcp.Client,
	logger *log.Logger,
	storage storage.ChatHistoryStorage,
	toolsProvider *goai.ToolsProvider,
) *chi.Mux {
	return setupRouter(mcpClient, logger, storage, toolsProvider)
}

func NewContainer(
	logger *log.Logger,
	mcpClient *mcp.Client,
	toolsProvider *goai.ToolsProvider,
	chatHistoryStorage storage.ChatHistoryStorage,
	router *chi.Mux,
	config *config.Config,
) *Container {
	return &Container{
		Logger:             logger,
		MCPClient:          mcpClient,
		ToolsProvider:      toolsProvider,
		ChatHistoryStorage: chatHistoryStorage,
		Router:             router,
		Config:             config,
	}
}
