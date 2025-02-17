package cmd

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"time"

	"log"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

type QuestionRequest struct {
	Question string `json:"question"`
	UseTools bool   `json:"useTools"`
}

type Response struct {
	Answer      string `json:"answer"`
	InputToken  int    `json:"input_token"`
	OutputToken int    `json:"output_token"`
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

			mcpClient := mcp.NewClient(mcp.NewSSETransport(), mcp.ClientConfig{
				ClientName:    "My MCP Kit Client",
				ClientVersion: "1.0.0",
				Logger:        log.New(logger.Writer(), "", log.LstdFlags),
				RetryDelay:    5 * time.Second,
				MaxRetries:    3,
				SSE: mcp.SSEConfig{
					URL: cfg.MCPServerURL,
				},
			})
			defer mcpClient.Close(ctx)

			if err := mcpClient.Connect(context.Background()); err != nil {
				log.Printf("Failed to connect to SSE: %v", err)
				return fmt.Errorf("failed to connect to MCP server: %w", err)
			}

			router := setupRouter(ctx, mcpClient, logger)
			logger.Printf("Starting server on :8080")
			return http.ListenAndServe(":8081", router)
		},
	}
}

func setupRouter(ctx context.Context, mcpClient *mcp.Client, logger *log.Logger) *chi.Mux {
	r := chi.NewRouter()

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

	// Handle other static files
	r.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFS))))

	r.Post("/ask", handleAsk(ctx, mcpClient, logger))

	return r
}

func handleAsk(ctx context.Context, sseClient *mcp.Client, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req QuestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		reqOptions := []goai.RequestOption{
			goai.WithMaxToken(1000),
			goai.WithTemperature(0.5),
		}

		if req.UseTools {
			logger.Println("Using tools provider")
			toolsProvider := goai.NewToolsProvider()
			if err := toolsProvider.AddMCPClient(sseClient); err != nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Failed to connect with MCP client",
				})
				return
			}
			reqOptions = append(reqOptions, goai.UseToolsProvider(toolsProvider))
		}

		llmProvider := goai.NewAnthropicLLMProvider(goai.AnthropicProviderConfig{
			Client: goai.NewAnthropicClient(os.Getenv("ANTHROPIC_API_KEY")),
			Model:  anthropic.ModelClaude3_5Sonnet20241022,
		})
		llm := goai.NewLLMRequest(goai.NewRequestConfig(reqOptions...), llmProvider)

		messages, err := buildMessagesFromPromptTemplates(ctx, sseClient, req)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		response, err := llm.Generate(ctx, messages)
		if err != nil {
			log.Printf("Failed to generate response: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to generate response",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			Answer:      response.Text,
			InputToken:  response.TotalInputToken,
			OutputToken: response.TotalOutputToken,
		})
	}
}

func buildMessagesFromPromptTemplates(ctx context.Context, sseClient *mcp.Client, req QuestionRequest) ([]goai.LLMMessage, error) {
	promptName := "llm_general"
	if req.UseTools {
		promptName = "llm_with_tools"
	}

	promptMessages, err := sseClient.GetPrompt(ctx, mcp.GetPromptParams{
		Name: promptName,
		Arguments: json.RawMessage(fmt.Sprintf(`{
			"question": "%s"
		}`, req.Question)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prompts from MCP server")
	}

	var messages []goai.LLMMessage
	for _, p := range promptMessages {
		messages = append(messages, goai.LLMMessage{
			Role: goai.LLMMessageRole(p.Role),
			Text: p.Content.Text,
		})
	}
	return messages, nil
}
