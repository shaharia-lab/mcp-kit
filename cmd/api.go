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
	"github.com/google/uuid"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	goaiObs "github.com/shaharia-lab/goai/observability"
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

type ModelSettings struct {
	Temperature float64 `json:"temperature"`
	MaxTokens   int64   `json:"maxTokens"`
	TopP        float64 `json:"topP"`
	TopK        int64   `json:"topK"`
}
type QuestionRequest struct {
	ChatUUID      uuid.UUID     `json:"chat_uuid"`
	Question      string        `json:"question"`
	UseTools      bool          `json:"useTools"`
	ModelSettings ModelSettings `json:"modelSettings"`
}

type Response struct {
	ChatUUID    uuid.UUID `json:"chat_uuid"`
	Answer      string    `json:"answer"`
	InputToken  int       `json:"input_token"`
	OutputToken int       `json:"output_token"`
}

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

	r.Get("/llm-providers", func(w http.ResponseWriter, r *http.Request) {
		providers := getLLMProviders()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(providers); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	// Handle other static files
	r.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFS))))

	r.Post("/ask", handleAsk(mcpClient, logger, chatHistoryStorage, toolsProvider))
	r.Get("/chats", handleListChats(logger, chatHistoryStorage))
	r.Get("/chat/{chatId}", handleGetChat(logger, chatHistoryStorage))
	r.Get("/api/tools", handleListTools(toolsProvider))

	return r
}

func handleAsk(sseClient *mcp.Client, logger *log.Logger, historyStorage storage.ChatHistoryStorage, toolsProvider *goai.ToolsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "handle_ask")
		defer span.End()

		var req QuestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			observability.AddAttribute(ctx, "error", err.Error())
			defer span.End()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid request body",
			})
			return
		}
		observability.AddAttribute(ctx, "question.length", len(req.Question))
		observability.AddAttribute(ctx, "question.use_tools", req.UseTools)

		var chat *storage.ChatHistory
		var err error

		// Check if ChatUUID is nil (zero ChatUUID)
		if req.ChatUUID == uuid.Nil {
			// Create new chat if ChatUUID is nil
			chat, err = historyStorage.CreateChat()
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Failed to create new chat",
				})
				return
			}
		} else {
			logger.Printf("ChatUUID: %s", req.ChatUUID)
			chat, err = historyStorage.GetChat(req.ChatUUID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Chat not found",
				})
				return
			}
		}

		// Add message to the chat
		err = historyStorage.AddMessage(chat.UUID, storage.Message{
			LLMMessage: goai.LLMMessage{
				Role: goai.UserRole,
				Text: req.Question,
			},
			GeneratedAt: time.Now(),
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to add message to chat history",
			})
			return
		}

		reqOptions := []goai.RequestOption{
			goai.WithMaxToken(1000),
			goai.WithTemperature(0.5),
		}

		if req.ModelSettings.Temperature != 0 {
			reqOptions = append(reqOptions, goai.WithTemperature(req.ModelSettings.Temperature))
		}

		log.Printf("MaxTokens: %d", req.ModelSettings.MaxTokens)
		if req.ModelSettings.MaxTokens != 0 {
			reqOptions = append(reqOptions, goai.WithMaxToken(req.ModelSettings.MaxTokens))
		}

		if req.ModelSettings.TopP != 0 {
			reqOptions = append(reqOptions, goai.WithTopP(req.ModelSettings.TopP))
		}

		if req.ModelSettings.TopK != 0 {
			reqOptions = append(reqOptions, goai.WithTopK(req.ModelSettings.TopK))
		}

		if req.UseTools {
			observability.AddAttribute(ctx, "tools.enabled", true)

			logger.Println("Using tools provider")
			reqOptions = append(reqOptions, goai.UseToolsProvider(toolsProvider))
		}

		llmSetupCtx, llmSetupSpan := observability.StartSpan(ctx, "setup_llm_provider")

		/*llmProvider := goai.NewAnthropicLLMProvider(goai.AnthropicProviderConfig{
			Client: goai.NewAnthropicClient(os.Getenv("ANTHROPIC_API_KEY")),
			Model:  anthropic.ModelClaude3_5Sonnet20241022,
		})
		llm := goai.NewLLMRequest(goai.NewRequestConfig(reqOptions...), llmProvider)
		*/
		observability.AddAttribute(llmSetupCtx, "llm.model", anthropic.ModelClaude3_5Sonnet20241022)
		llmSetupSpan.End()

		// Span for building messages
		messagesCtx, messagesSpan := observability.StartSpan(ctx, "build_messages")
		messages, err := buildMessagesFromPromptTemplates(messagesCtx, sseClient, req)

		if err != nil {
			observability.AddAttribute(messagesCtx, "error", err.Error())
			messagesSpan.End()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}
		observability.AddAttribute(messagesCtx, "messages.count", len(messages))
		messagesSpan.End()

		generateCtx, generateSpan := observability.StartSpan(ctx, "generate_response")
		//response, err := llm.Generate(generateCtx, messages)
		response := goai.LLMResponse{
			Text:             "# Hello! 👋\n\nI'm your AI assistant, ready to help you. To provide the most useful assistance, I can help you with:\n\n- Answering questions\n- Solving problems\n- Explaining concepts\n- Writing and reviewing code\n- General discussion and information\n\n## How can I help you today?\n\nPlease feel free to ask any specific question or let me know what kind of assistance you need. I'll make sure to:\n1. Understand your request clearly\n2. Ask for any needed clarification\n3. Provide well-formatted, helpful responses",
			TotalInputToken:  10,
			TotalOutputToken: 10,
			CompletionTime:   2,
		}

		if err != nil {
			observability.AddAttribute(generateCtx, "error", err.Error())
			generateSpan.End()

			log.Printf("Failed to generate response: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to generate response",
			})
			return
		}

		err = historyStorage.AddMessage(chat.UUID, storage.Message{
			LLMMessage: goai.LLMMessage{
				Role: goai.SystemRole,
				Text: response.Text,
			},
			GeneratedAt: time.Now(),
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to add message to chat history",
			})
			return
		}

		observability.AddAttribute(generateCtx, "response.input_tokens", response.TotalInputToken)
		observability.AddAttribute(generateCtx, "response.output_tokens", response.TotalOutputToken)
		generateSpan.End()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			ChatUUID:    chat.UUID,
			Answer:      response.Text,
			InputToken:  response.TotalInputToken,
			OutputToken: response.TotalOutputToken,
		})
	}
}

func handleListChats(logger *log.Logger, historyStorage storage.ChatHistoryStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chats, err := historyStorage.ListChatHistories()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to retrieve chat histories",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := struct {
			Chats []storage.ChatHistory `json:"chats"`
		}{
			Chats: chats,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to encode response",
			})
			return
		}
	}
}

func handleGetChat(logger *log.Logger, historyStorage storage.ChatHistoryStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract chat UUID from URL
		chatUUID := chi.URLParam(r, "chatId")
		if chatUUID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Chat ID is required",
			})
			return
		}

		parsedChatUUID, err := uuid.Parse(chatUUID)
		if err != nil {
			logger.Printf("Failed to parse chat UUID: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid chat ID",
			})
			return
		}

		// Get chat history from storage
		chat, err := historyStorage.GetChat(parsedChatUUID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Chat not found",
			})
			return
		}

		// Return the chat history
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(chat); err != nil {
			logger.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to encode response",
			})
			return
		}
	}
}

func handleListTools(toolsProvider *goai.ToolsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "handle_list_tools")
		defer span.End()

		tools, err := toolsProvider.ListTools(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get tools: %v", err), http.StatusInternalServerError)
			span.RecordError(err)
			return
		}

		// Convert to simplified format
		toolInfos := make([]ToolInfo, len(tools))
		for i, tool := range tools {
			toolInfos[i] = ToolInfo{
				Name:        tool.Name,
				Description: tool.Description,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(toolInfos); err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
			span.RecordError(err)
			return
		}
	}
}

func buildMessagesFromPromptTemplates(ctx context.Context, sseClient *mcp.Client, req QuestionRequest) ([]goai.LLMMessage, error) {
	promptName := "llm_general"
	if req.UseTools {
		promptName = "llm_with_tools"
	}
	log.Printf("Fetching prompt: %s", promptName)

	promptMessages, err := sseClient.GetPrompt(ctx, mcp.GetPromptParams{
		Name: promptName,
		Arguments: json.RawMessage(`{
        "question": ` + fmt.Sprintf("%q", req.Question) + `
    }`),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch prompts from MCP server. Error: %w", err)
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

// Model represents an LLM model's information
type Model struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ModelID     string `json:"modelId"`
}

// Provider represents an LLM provider and its available models
type Provider struct {
	Name   string  `json:"name"`
	Models []Model `json:"models"`
}

// SupportedLLMProviders represents the response structure for the API endpoint
type SupportedLLMProviders struct {
	Providers []Provider `json:"providers"`
}

// getLLMProviders returns a list of supported LLM providers and their models
func getLLMProviders() SupportedLLMProviders {
	return SupportedLLMProviders{
		Providers: []Provider{
			{
				Name: "Anthropic",
				Models: []Model{
					{
						Name:        "Claude 3 Sonnet",
						Description: "Advanced language model optimized for reliability and thoughtful responses",
						ModelID:     "claude-3-sonnet-20240229",
					},
					{
						Name:        "Claude 2.1",
						Description: "Powerful language model for general-purpose tasks",
						ModelID:     "claude-2.1",
					},
				},
			},
			{
				Name: "OpenAI",
				Models: []Model{
					{
						Name:        "GPT-4 Turbo",
						Description: "Most capable GPT-4 model for various tasks",
						ModelID:     "gpt-4-turbo-preview",
					},
					{
						Name:        "GPT-3.5 Turbo",
						Description: "Efficient model balancing performance and speed",
						ModelID:     "gpt-3.5-turbo",
					},
				},
			},
		},
	}
}
