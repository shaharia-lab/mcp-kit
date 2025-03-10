package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shaharia-lab/mcp-kit/internal/service/google"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/mcp-kit/internal/auth"
	handlers "github.com/shaharia-lab/mcp-kit/internal/handler"
	"github.com/shaharia-lab/mcp-kit/internal/observability"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"log"
	"net/http"
)

// ToolInfo represents the simplified tool information to be returned by the API
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RouterDependencies struct {
	MCPClient          *mcp.Client
	Logger             *log.Logger
	ChatHistoryStorage goai.ChatHistoryStorage
	ToolsProvider      *goai.ToolsProvider
}

func NewAPICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Start the API server",
		Long:  "Start the API server with LLM endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Create context with cancellation for graceful shutdown
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			// Setup signal handling for graceful shutdown
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

			// Initialize all dependencies using Wire
			container, cleanup, err := InitializeAPI(ctx, configFile)
			if err != nil {
				return fmt.Errorf("failed to initialize application: %w", err)
			}
			defer cleanup()

			// Initialize the tracing service
			if err = container.TracingService.Initialize(ctx); err != nil {
				return fmt.Errorf("failed to initialize tracer: %w", err)
			}

			defer func() {
				if err := container.TracingService.Shutdown(ctx); err != nil {
					container.Logger.Printf("Error shutting down tracer: %v", err)
				}
			}()

			tracer := otel.Tracer("mcp-kit")

			// Connect MCP client with span
			clientCtx, clientSpan := tracer.Start(ctx, "mcp_client_init")
			if err := container.MCPClient.Connect(clientCtx); err != nil {
				clientSpan.End()
				return fmt.Errorf("failed to connect to MCP server: %w", err)
			}
			clientSpan.End()

			// Create HTTP server
			srv := &http.Server{
				Addr: fmt.Sprintf(":%d", container.Config.APIServerPort),
				Handler: setupRouter(
					container.MCPClient,
					container.Logger,
					container.ChatHistoryStorage,
					container.ToolsProvider,
					container.AuthMiddleware,
					container.GoogleService,
				),
			}

			// Channel to capture server errors
			serverErrors := make(chan error, 1)

			// Start server in a goroutine
			go func() {
				container.Logger.Printf("Starting server on %s", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					serverErrors <- err
				}
			}()

			// Wait for shutdown signal or server error
			select {
			case err = <-serverErrors:
				return fmt.Errorf("server error: %w", err)
			case sig := <-signalChan:
				container.Logger.Printf("Received signal: %v", sig)

				// Create shutdown context with timeout
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer shutdownCancel()

				// First disconnect MCP client
				container.Logger.Printf("Disconnecting MCP client...")
				if err = container.MCPClient.Close(shutdownCtx); err != nil {
					container.Logger.Printf("Error disconnecting MCP client: %v", err)
				}

				// Then shutdown the HTTP server
				container.Logger.Printf("Shutting down HTTP server...")
				if err = srv.Shutdown(shutdownCtx); err != nil {
					return fmt.Errorf("server shutdown error: %w", err)
				}

				container.Logger.Printf("Server shutdown completed")
				return nil
			}
		},
	}
}

func setupRouter(
	mcpClient *mcp.Client,
	logger *log.Logger,
	chatHistoryStorage goai.ChatHistoryStorage,
	toolsProvider *goai.ToolsProvider,
	authMiddleware *auth.AuthMiddleware,
	googleService *google.GoogleService,
) *chi.Mux {
	r := chi.NewRouter()

	// Tracing middleware remains the same
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := otel.Tracer("http-middleware").Start(r.Context(), "http_request")
			defer span.End()

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.path", r.URL.Path),
				attribute.String("http.host", r.Host),
			)

			w = wrapResponseWriter(w, span)

			next.ServeHTTP(w, r.WithContext(ctx))
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

	r.Use(prometheusMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Expose the metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ping": "Pong"}`))
	})

	// Root route redirects to /static
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static", http.StatusFound)
	})

	r.Get("/llm-providers", handlers.LLMProvidersHandler)

	r.Post("/ask", handlers.HandleAsk(mcpClient, logger, chatHistoryStorage, toolsProvider))
	r.Post("/ask-stream", handlers.HandleAskStream(mcpClient, logger, chatHistoryStorage, toolsProvider))
	r.With(authMiddleware.EnsureValidToken).Get("/chats", handlers.ChatHistoryListsHandler(logger, chatHistoryStorage))
	r.Get("/chats/{chatId}", handlers.GetChatHandler(logger, chatHistoryStorage))
	r.Get("/api/tools", handlers.ListToolsHandler(toolsProvider))

	r.Get("/oauth/login", func(w http.ResponseWriter, r *http.Request) {
		googleService.HandleOAuthStart(w, r)
	})

	r.Get("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		googleService.HandleOAuthCallback(w, r)
		logger.Printf("Authenticate with Google Service has been successfully completed")
		w.Write([]byte("Authentication successful. You can close this window now."))
	})

	return r
}

// Prometheus middleware to collect metrics
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			// Record the request duration
			duration := time.Since(start).Seconds()
			observability.HTTPRequestDuration.WithLabelValues(r.URL.Path).Observe(duration)

			// Record the request count
			observability.HTTPRequestsTotal.WithLabelValues(
				fmt.Sprintf("%d", ww.Status()),
				r.Method,
			).Inc()
		}()

		next.ServeHTTP(ww, r)
	})
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
	rw.span.SetAttributes(attribute.Int("http.status_code", code))
	rw.ResponseWriter.WriteHeader(code)
}

// Flush Add this method to allow flushing through the wrapper
func (rw *responseWriterWrapper) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
