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

				// Create separate contexts for MCP client and HTTP server shutdown
				mcpShutdownCtx, mcpShutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer mcpShutdownCancel()

				// Use a channel to handle MCP client shutdown timeout
				mcpClosed := make(chan struct{})
				go func() {
					container.Logger.Printf("Disconnecting MCP client...")
					if err = container.MCPClient.Close(mcpShutdownCtx); err != nil {
						container.Logger.Printf("Error disconnecting MCP client: %v", err)
					}
					close(mcpClosed)
				}()

				// Wait for MCP client to close or timeout
				select {
				case <-mcpClosed:
					container.Logger.Printf("MCP client disconnected successfully")
				case <-mcpShutdownCtx.Done():
					container.Logger.Printf("MCP client disconnect timed out, proceeding with server shutdown")
				}

				// Create separate context for HTTP server shutdown
				serverShutdownCtx, serverShutdownCancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer serverShutdownCancel()

				// Then shutdown the HTTP server
				container.Logger.Printf("Shutting down HTTP server...")
				if err = srv.Shutdown(serverShutdownCtx); err != nil {
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
		ExposedHeaders:   []string{"Link", "X-MKit-Chat-UUID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(prometheusMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	sunsetDate := "Sat March 31 2025 23:59:59 GMT"

	// Deprecated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.NoCache)

		r.With(DeprecatedRouteMiddleware(DeprecationInfo{
			SuccessorURL: "/api/v1/llm-providers",
			SunsetDate:   sunsetDate,
		})).Get("/llm-providers", handlers.LLMProvidersHandler)

		r.With(DeprecatedRouteMiddleware(DeprecationInfo{
			SuccessorURL: "/api/v1/chats/{chatId}",
			SunsetDate:   sunsetDate,
		})).Get("/chats/{chatId}", handlers.GetChatHandler(logger, chatHistoryStorage))

		r.With(DeprecatedRouteMiddleware(DeprecationInfo{
			SuccessorURL: "/api/v1/tools",
			SunsetDate:   sunsetDate,
		})).Get("/api/tools", handlers.ListToolsHandler(toolsProvider))

		r.With(DeprecatedRouteMiddleware(DeprecationInfo{
			SuccessorURL: "/api/v1/chats",
			SunsetDate:   sunsetDate,
		})).Post("/ask", handlers.HandleAsk(mcpClient, logger, chatHistoryStorage, toolsProvider))

		r.With(DeprecatedRouteMiddleware(DeprecationInfo{
			SuccessorURL: "/api/v1/chats/stream",
			SunsetDate:   sunsetDate,
		})).Post("/ask-stream", handlers.HandleAskStream(mcpClient, logger, chatHistoryStorage, toolsProvider))

		r.With(
			authMiddleware.EnsureValidToken,
			DeprecatedRouteMiddleware(DeprecationInfo{
				SuccessorURL: "/api/v1/chats",
				SunsetDate:   sunsetDate,
			}),
		).Get("/chats", handlers.ChatHistoryListsHandler(logger, chatHistoryStorage))
	})

	// Expose the metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ping": "Pong"}`))
	})

	// Get the list of LLM providers
	r.Route("/api/v1/llm-providers", func(r chi.Router) {
		r.Use(authMiddleware.EnsureValidToken)
		r.Get("/", handlers.LLMProvidersHandler)
	})

	// Get the list of tools
	r.Route("/api/v1/tools", func(r chi.Router) {
		r.Use(authMiddleware.EnsureValidToken)
		r.Get("/", handlers.ListToolsHandler(toolsProvider))
	})

	// Ask LLM a question, with or without streaming
	// Also get the chat history
	r.Route("/api/v1/chats", func(r chi.Router) {
		r.Use(authMiddleware.EnsureValidToken)
		r.Post("/", handlers.HandleAsk(mcpClient, logger, chatHistoryStorage, toolsProvider))
		r.Post("/stream", handlers.HandleAskStream(mcpClient, logger, chatHistoryStorage, toolsProvider))
		r.Get("/{chatId}", handlers.GetChatHandler(logger, chatHistoryStorage))
		r.Get("/", handlers.ChatHistoryListsHandler(logger, chatHistoryStorage))
	})

	// Authenticate with Google OAuth2 to access Google services like Gmail Tools
	r.Route("/google-oauth2", func(r chi.Router) {
		r.With(authMiddleware.EnsureValidToken).Get("/login", func(w http.ResponseWriter, r *http.Request) {
			googleService.HandleOAuthStart(w, r)
		})

		r.Get("/callback", func(w http.ResponseWriter, r *http.Request) {
			googleService.HandleOAuthCallback(w, r)
			logger.Printf("Authenticate with Google Service has been successfully completed")
			w.Write([]byte("Authentication successful. You can close this window now."))
		})
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

// DeprecationInfo holds information about a deprecated endpoint
type DeprecationInfo struct {
	SuccessorURL string
	SunsetDate   string
}

// DeprecatedRouteMiddleware creates a middleware that adds deprecation headers
func DeprecatedRouteMiddleware(info DeprecationInfo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Deprecation", "true")
			w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"successor-version\"", info.SuccessorURL))
			w.Header().Set("Sunset", info.SunsetDate)

			// Optional: Add custom header for more details
			w.Header().Set("X-API-Deprecated-Message",
				fmt.Sprintf("This endpoint is deprecated. Please use %s instead", info.SuccessorURL))

			next.ServeHTTP(w, r)
		})
	}
}
