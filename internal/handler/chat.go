package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/mcp-kit/internal/observability"
	"github.com/shaharia-lab/mcp-kit/internal/service/llm"
)

const (
	MaxChatHistoryMessages = 20 // Configurable max number of messages to retain
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
	SelectedTools []string      `json:"selectedTools"`
	ModelSettings ModelSettings `json:"modelSettings"`
	LLMProvider   LLMProvider   `json:"llmProvider"`
}

type Response struct {
	ChatUUID    uuid.UUID `json:"chat_uuid"`
	Answer      string    `json:"answer"`
	InputToken  int       `json:"input_token"`
	OutputToken int       `json:"output_token"`
}

func HandleAsk(sseClient *mcp.Client, logger *log.Logger, historyStorage goai.ChatHistoryStorage, toolsProvider *goai.ToolsProvider) http.HandlerFunc {
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

		supportedLLMProviders := getLLMProviders()
		if supportedLLMProviders.IsSupported(req.LLMProvider.Provider, req.LLMProvider.ModelID) == false {
			writeErrorResponse(w, http.StatusBadRequest, "LLM provider or model is not supported", nil, ctx, span)
			return
		}

		observability.AddAttribute(ctx, "question.length", len(req.Question))
		observability.AddAttribute(ctx, "question.use_tools", req.SelectedTools)
		observability.AddAttribute(ctx, "model.temperature", req.ModelSettings.Temperature)
		observability.AddAttribute(ctx, "model.max_tokens", req.ModelSettings.MaxTokens)
		observability.AddAttribute(ctx, "model.top_p", req.ModelSettings.TopP)
		observability.AddAttribute(ctx, "model.top_k", req.ModelSettings.TopK)
		observability.AddAttribute(ctx, "llm.provider", req.LLMProvider.Provider)
		observability.AddAttribute(ctx, "llm.model_id", req.LLMProvider.ModelID)

		// Retrieve or create chat history
		chat, err := getOrInitializeChat(req, historyStorage, logger, ctx)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to get or create chat history", err, ctx, span)
			return
		}

		// Retrieve chat history and truncate it to the last MaxChatHistoryMessages
		messages, err := getTruncatedChatHistory(ctx, chat.UUID, historyStorage)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve chat history", err, ctx, span)
			return
		}

		// Build the initial prompt template if this is the first message in the chat
		if len(messages) == 0 {
			promptMessages, err := buildMessagesFromPromptTemplates(ctx, sseClient, req)
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Failed to build prompt templates", err, ctx, span)
				return
			}
			messages = append(messages, promptMessages...)
		}

		// Add the new user question to the history
		userMessage := goai.LLMMessage{
			Role: goai.UserRole,
			Text: req.Question,
		}
		messages = append(messages, userMessage)

		// Add user message to chat history
		if err := historyStorage.AddMessage(ctx, chat.UUID, goai.ChatHistoryMessage{
			LLMMessage:  userMessage,
			GeneratedAt: time.Now(),
		}); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to add user message to chat history", err, ctx, span)
			return
		}

		// Prepare options for the LLM request
		reqOptions := prepareLLMRequestOptions(req)

		if len(req.SelectedTools) > 0 {
			logger.Println("Using tools provider")
			reqOptions = append(reqOptions, goai.UseToolsProvider(toolsProvider))
			observability.AddAttribute(ctx, "tools.enabled", true)
			reqOptions = append(reqOptions, goai.WithAllowedTools(req.SelectedTools))

			observability.ToolsEnabledTotal.WithLabelValues(req.LLMProvider.Provider, req.LLMProvider.ModelID).Inc()
		}

		for _, tool := range req.SelectedTools {
			observability.ToolsUsageTotal.WithLabelValues(
				tool,
				req.LLMProvider.Provider,
				req.LLMProvider.ModelID,
			).Inc()
		}

		builder := llm.NewLLMBuilder(ctx)
		llmProvider, err := builder.BuildProvider(llm.ProviderConfig{
			Provider: req.LLMProvider.Provider,
			ModelID:  req.LLMProvider.ModelID,
		})
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil, ctx, span)
			return
		}

		llmCompletion := goai.NewLLMRequest(goai.NewRequestConfig(reqOptions...), llmProvider)

		// Generate a response using the LLM
		generateCtx, generateSpan := observability.StartSpan(ctx, "generate_response")
		observability.AddAttribute(generateCtx, "HandleAsk.total_messages", len(messages))

		timer := prometheus.NewTimer(observability.LLMCompletionDuration.WithLabelValues(
			req.LLMProvider.Provider,
			req.LLMProvider.ModelID,
			"success",
		))
		defer timer.ObserveDuration()

		// Track in-flight requests
		inFlightMetric := observability.LLMCompletionInFlight.WithLabelValues(
			req.LLMProvider.Provider,
			req.LLMProvider.ModelID,
		)
		inFlightMetric.Inc()
		defer inFlightMetric.Dec()

		response, err := llmCompletion.Generate(generateCtx, messages)
		if err != nil {
			// Record failed completion duration with "error" status
			observability.LLMCompletionDuration.WithLabelValues(
				req.LLMProvider.Provider,
				req.LLMProvider.ModelID,
				"error",
			).Observe(timer.ObserveDuration().Seconds())

			// Your existing error handling code...
			observability.AddAttribute(generateCtx, "error", err.Error())
			generateSpan.End()
			log.Printf("Failed to generate response: %v", err)
			writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate response. error: %s", err.Error()), err, generateCtx, span)
			return
		}

		generateSpan.End()

		observability.TokensInputTotal.WithLabelValues(req.LLMProvider.Provider, req.LLMProvider.ModelID).
			Add(float64(response.TotalInputToken))
		observability.TokensOutputTotal.WithLabelValues(req.LLMProvider.Provider, req.LLMProvider.ModelID).
			Add(float64(response.TotalOutputToken))

		// Add assistant response to chat history
		err = historyStorage.AddMessage(ctx, chat.UUID, goai.ChatHistoryMessage{
			LLMMessage: goai.LLMMessage{
				Role: goai.AssistantRole,
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
func getOrInitializeChat(req QuestionRequest, historyStorage goai.ChatHistoryStorage, logger *log.Logger, ctx context.Context) (*goai.ChatHistory, error) {
	var chat *goai.ChatHistory
	var err error

	if req.ChatUUID == uuid.Nil {
		chat, err = historyStorage.CreateChat(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create chat")
		}
	} else {
		logger.Printf("ChatUUID: %s", req.ChatUUID)
		chat, err = historyStorage.GetChat(ctx, req.ChatUUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get chat")
		}
	}

	return chat, nil
}

func getTruncatedChatHistory(ctx context.Context, chatUUID uuid.UUID, historyStorage goai.ChatHistoryStorage) ([]goai.LLMMessage, error) {
	// Retrieve all messages for the chat
	chatHistory, err := historyStorage.GetChat(ctx, chatUUID)
	if err != nil {
		return nil, err
	}

	if chatHistory == nil {
		return []goai.LLMMessage{}, nil
	}

	messages := chatHistory.Messages
	var result []goai.LLMMessage

	// Determine the starting index for truncation
	startIdx := 0
	if len(messages) > MaxChatHistoryMessages {
		startIdx = len(messages) - MaxChatHistoryMessages
	}

	// Ensure System messages are always included
	for _, msg := range messages {
		if msg.Role == goai.SystemRole {
			result = append(result, goai.LLMMessage{
				Role: msg.Role,
				Text: msg.Text,
			})
		}
	}

	// Convert Message to goai.LLMMessage
	for _, msg := range messages[startIdx:] {
		if msg.Role != goai.SystemRole {
			llmMsg := goai.LLMMessage{
				Role: msg.Role,
				Text: msg.Text,
			}
			result = append(result, llmMsg)
		}
	}

	return result, nil
}

func buildMessagesFromPromptTemplates(ctx context.Context, sseClient *mcp.Client, req QuestionRequest) ([]goai.LLMMessage, error) {
	promptName := "llm_general"
	if len(req.SelectedTools) > 0 {
		promptName = "llm_with_tools_v2"
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
