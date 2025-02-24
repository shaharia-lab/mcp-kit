package cmd

import (
	"context"
	"embed"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	goaiObs "github.com/shaharia-lab/goai/observability"
	handlers "github.com/shaharia-lab/mcp-kit/internal/handler"
	"github.com/shaharia-lab/mcp-kit/observability"
	"github.com/shaharia-lab/mcp-kit/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/trace"
	"io/fs"
	"time"

	"log"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

// ToolInfo represents the simplified tool information to be returned by the API
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewAPICmd(logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Start the API server",
		Long:  "Start the API server with LLM endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			ctx := context.Background()
			cleanup, err := initializeTracer(ctx, "mcp-kit", logrus.New())
			if err != nil {
				return err
			}
			defer cleanup()

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			ctx, serverSpan := observability.StartSpan(ctx, "server_lifecycle")
			defer serverSpan.End()

			l := goaiObs.NewDefaultLogger()

			// Add span for MCP client initialization
			clientCtx, clientSpan := observability.StartSpan(ctx, "mcp_client_init")

			mcpClient := mcp.NewClient(mcp.NewSSETransport(l), mcp.ClientConfig{
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
			defer mcpClient.Close(ctx)

			if err := mcpClient.Connect(clientCtx); err != nil {
				clientSpan.End()
				log.Printf("Failed to connect to SSE: %v", err)
				return fmt.Errorf("failed to connect to MCP server: %w", err)
			}
			clientSpan.End()

			inMemoryChatHistoryStorage := storage.NewInMemoryChatHistoryStorage()

			toolsProvider := goai.NewToolsProvider()
			if err := toolsProvider.AddMCPClient(mcpClient); err != nil {
				log.Printf("Failed to add MCP client to tools provider: %v", err)
				return fmt.Errorf("failed to add MCP client to tools provider: %w", err)
			}

			router := setupRouter(ctx, mcpClient, logger, inMemoryChatHistoryStorage, toolsProvider)
			logger.Printf("Starting server on :8080")

			observability.AddAttribute(ctx, "server.port", "8081")
			return http.ListenAndServe(":8081", router)
		},
	}
}

func setupRouter(ctx context.Context, mcpClient *mcp.Client, logger *log.Logger, chatHistoryStorage storage.ChatHistoryStorage, toolsProvider *goai.ToolsProvider) *chi.Mux {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCtx, span := observability.StartSpan(r.Context(), "http_request")
			defer span.End()

			observability.AddAttribute(requestCtx, "http.method", r.Method)
			observability.AddAttribute(requestCtx, "http.url", r.URL.String())
			observability.AddAttribute(requestCtx, "http.path", r.URL.Path)
			observability.AddAttribute(requestCtx, "http.host", r.Host)

			w = wrapResponseWriter(w, span)

			next.ServeHTTP(w, r.WithContext(requestCtx))

		})
	})

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Root route redirects to /static
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static", http.StatusFound)
	})

	// Create a sub-filesystem to handle the static directory correctly
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Handle /static route to serve index.html
	r.Get("/static", func(w http.ResponseWriter, r *http.Request) {
		content, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			http.Error(w, "Index file not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(content)
	})

	r.Get("/llm-providers", handlers.LLMProvidersHandler)

	// Handle other static files
	r.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFS))))

	r.Post("/ask", handlers.HandleAsk(mcpClient, logger, chatHistoryStorage, toolsProvider))
	r.Get("/chats", handlers.ChatHistoryListsHandler(logger, chatHistoryStorage))
	r.Get("/chat/{chatId}", handlers.GetChatHandler(logger, chatHistoryStorage))
	r.Get("/api/tools", handlers.ListToolsHandler(toolsProvider))

	return r
}

func initializeTracer(ctx context.Context, appName string, l *logrus.Logger) (func(), error) {
	l.Info("Initializing tracer")
	cleanup, err := observability.InitTracer(ctx, appName, l)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
		return nil, err
	}
	l.Info("Tracer initialized successfully")
	return cleanup, nil
}

// Helper to track response status
type responseWriterWrapper struct {
	http.ResponseWriter
	status int
	span   trace.Span
}

func wrapResponseWriter(w http.ResponseWriter, span trace.Span) *responseWriterWrapper {
	return &responseWriterWrapper{w, http.StatusOK, span}
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.status = code
	observability.AddAttribute(context.Background(), "http.status_code", code)
	rw.ResponseWriter.WriteHeader(code)
}
