package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

type QuestionRequest struct {
	Question string `json:"question"`
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

			sseClient, err := initializeSSEClient(cfg, logger)
			if err != nil {
				return fmt.Errorf("failed to initialize SSE client: %w", err)
			}
			defer sseClient.Close(ctx)

			router := setupRouter(ctx, sseClient, logger)
			logger.Printf("Starting server on :8080")
			return http.ListenAndServe(":8081", router)
		},
	}
}

func setupRouter(ctx context.Context, mcpClient *mcp.Client, logger *log.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Post("/ask", handleAsk(ctx, mcpClient, logger))

	return r
}

func handleAsk(ctx context.Context, sseClient *mcp.Client, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req QuestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		llm, err := initializeLLM(sseClient)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response, err := llm.Generate(ctx, []goai.LLMMessage{
			{Role: goai.UserRole, Text: req.Question},
		})
		if err != nil {
			http.Error(w, "Failed to generate response", http.StatusInternalServerError)
			return
		}

		apiResponse := Response{
			Answer:      response.Text,
			InputToken:  response.TotalInputToken,
			OutputToken: response.TotalOutputToken,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiResponse)
	}
}
