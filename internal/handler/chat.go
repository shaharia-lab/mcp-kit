package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shaharia-lab/mcp-kit/internal/observability"
	"github.com/shaharia-lab/mcp-kit/internal/storage"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
)

type ModelSettings struct {
	Temperature float64 `json:"temperature"`
	MaxTokens   int64   `json:"maxTokens"`
	TopP        float64 `json:"topP"`
	TopK        int64   `json:"topK"`
}

type LLMProvider struct {
	Provider string `json:"provider"`
	ModelID  string `json:"modelId"`
}

type QuestionRequest struct {
	ChatUUID      uuid.UUID     `json:"chat_uuid"`
	Question      string        `json:"question"`
	UseTools      bool          `json:"useTools"`
	ModelSettings ModelSettings `json:"modelSettings"`
	LLMProvider   LLMProvider   `json:"llmProvider"`
}

type Response struct {
	ChatUUID    uuid.UUID `json:"chat_uuid"`
	Answer      string    `json:"answer"`
	InputToken  int       `json:"input_token"`
	OutputToken int       `json:"output_token"`
}

func HandleAsk(sseClient *mcp.Client, logger *log.Logger, historyStorage storage.ChatHistoryStorage, toolsProvider *goai.ToolsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "handle_ask")
		defer span.End()

		// Decode the incoming question request
		var req QuestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err, ctx, span)
			return
		}

		if req.Question == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Question cannot be empty", nil, ctx, span)
			return
		}

		if req.LLMProvider.Provider == "" || req.LLMProvider.ModelID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "LLM provider is required", nil, ctx, span)
			return
		}

		observability.AddAttribute(ctx, "question.length", len(req.Question))
		observability.AddAttribute(ctx, "question.use_tools", req.UseTools)

		// Retrieve or create chat history
		chat, err := getOrInitializeChat(req, historyStorage, logger, w, ctx)
		if err != nil {
			return // Error response is handled inside the function
		}

		// Add user message to chat history
		if err := addUserMessageToHistory(req.Question, chat.UUID, historyStorage, w); err != nil {
			return // Error response is handled inside the function
		}

		// Prepare options for the LLM request
		reqOptions := prepareLLMRequestOptions(req)

		if req.UseTools {
			logger.Println("Using tools provider")
			reqOptions = append(reqOptions, goai.UseToolsProvider(toolsProvider))
			observability.AddAttribute(ctx, "tools.enabled", true)
		}

		// Setup the LLM provider (Anthropic example)
		/*llmSetupCtx, llmSetupSpan := observability.StartSpan(ctx, "setup_llm_provider")

		// Uncomment this code if using an Anthropic provider:
		llmProvider := goai.NewAnthropicLLMProvider(goai.AnthropicProviderConfig{
			Client: goai.NewAnthropicClient(os.Getenv("ANTHROPIC_API_KEY")),
			Model:  anthropic.ModelClaude3_5Sonnet20241022,
		})
		llm := goai.NewLLMRequest(goai.NewRequestConfig(reqOptions...), llmProvider)

		observability.AddAttribute(llmSetupCtx, "llm.model", anthropic.ModelClaude3_5Sonnet20241022)
		llmSetupSpan.End()*/

		// Span for building messages
		messagesCtx, messagesSpan := observability.StartSpan(ctx, "build_messages")
		messages, err := buildMessagesFromPromptTemplates(messagesCtx, sseClient, req)
		if err != nil {
			observability.AddAttribute(messagesCtx, "error", err.Error())
			messagesSpan.End()
			writeErrorResponse(w, http.StatusInternalServerError, err.Error(), err, ctx, nil)
			return
		}
		observability.AddAttribute(messagesCtx, "messages.count", len(messages))
		messagesSpan.End()

		// Generate a response using the LLM
		/*generateCtx, generateSpan := observability.StartSpan(ctx, "generate_response")
		response, err := llm.Generate(generateCtx, messages)
		if err != nil {
			observability.AddAttribute(generateCtx, "error", err.Error())
			generateSpan.End()

			log.Printf("Failed to generate response: %v", err)
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate response", err, generateCtx, span)
			return
		}
		generateSpan.End()*/

		response := goai.LLMResponse{
			Text:             "# Hello! 👋\n\nI'm your AI assistant, ready to help you. To provide the most useful assistance, I can help you with:\n\n- Answering questions\n- Solving problems\n- Explaining concepts\n- Writing and reviewing code\n- General discussion and information\n\n## How can I help you today?\n\nPlease feel free to ask any specific question or let me know what kind of assistance you need. I'll make sure to:\n1. Understand your request clearly\n2. Ask for any needed clarification\n3. Provide well-formatted, helpful responses",
			TotalInputToken:  10,
			TotalOutputToken: 10,
			CompletionTime:   2,
		}

		// Add assistant response to chat history
		err = historyStorage.AddMessage(chat.UUID, storage.Message{
			LLMMessage: goai.LLMMessage{
				Role: goai.SystemRole,
				Text: response.Text,
			},
			GeneratedAt: time.Now(),
		})
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to add message to chat history", err, ctx, span)
			return
		}

		// Add attributes for observability
		observability.AddAttribute(ctx, "response.input_tokens", response.TotalInputToken)
		observability.AddAttribute(ctx, "response.output_tokens", response.TotalOutputToken)

		// Return the successful response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			ChatUUID:    chat.UUID,
			Answer:      response.Text,
			InputToken:  response.TotalInputToken,
			OutputToken: response.TotalOutputToken,
		})
	}
}

// Helper Function Implementations
func getOrInitializeChat(req QuestionRequest, historyStorage storage.ChatHistoryStorage, logger *log.Logger, w http.ResponseWriter, ctx context.Context) (*storage.ChatHistory, error) {
	var chat *storage.ChatHistory
	var err error

	if req.ChatUUID == uuid.Nil {
		chat, err = historyStorage.CreateChat()
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to create new chat", err, ctx, nil)
			return nil, err
		}
	} else {
		logger.Printf("ChatUUID: %s", req.ChatUUID)
		chat, err = historyStorage.GetChat(req.ChatUUID)
		if err != nil {
			writeErrorResponse(w, http.StatusNotFound, "Chat not found", err, ctx, nil)
			return nil, err
		}
	}

	return chat, nil
}

func addUserMessageToHistory(question string, chatUUID uuid.UUID, historyStorage storage.ChatHistoryStorage, w http.ResponseWriter) error {
	err := historyStorage.AddMessage(chatUUID, storage.Message{
		LLMMessage: goai.LLMMessage{
			Role: goai.UserRole,
			Text: question,
		},
		GeneratedAt: time.Now(),
	})
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to add message to chat history", err, nil, nil)
	}
	return err
}

func prepareLLMRequestOptions(req QuestionRequest) []goai.RequestOption {
	reqOptions := []goai.RequestOption{
		goai.WithMaxToken(1000),
		goai.WithTemperature(0.5),
	}

	if req.ModelSettings.Temperature != 0 {
		reqOptions = append(reqOptions, goai.WithTemperature(req.ModelSettings.Temperature))
	}

	if req.ModelSettings.MaxTokens != 0 {
		reqOptions = append(reqOptions, goai.WithMaxToken(req.ModelSettings.MaxTokens))
	}

	if req.ModelSettings.TopP != 0 {
		reqOptions = append(reqOptions, goai.WithTopP(req.ModelSettings.TopP))
	}

	if req.ModelSettings.TopK != 0 {
		reqOptions = append(reqOptions, goai.WithTopK(req.ModelSettings.TopK))
	}

	return reqOptions
}

func writeErrorResponse(w http.ResponseWriter, status int, message string, err error, ctx context.Context, span interface{}) {
	if err != nil {
		observability.AddAttribute(ctx, "error", err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
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
