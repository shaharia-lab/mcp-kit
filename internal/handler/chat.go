package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"

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

type StreamSettings struct {
	ChunkSize int `json:"chunk_size"`
	DelayMs   int `json:"delay_ms"`
}

type QuestionRequest struct {
	ChatUUID       uuid.UUID      `json:"chat_uuid"`
	Question       string         `json:"question"`
	SelectedTools  []string       `json:"selectedTools"`
	ModelSettings  ModelSettings  `json:"modelSettings"`
	LLMProvider    LLMProvider    `json:"llmProvider"`
	StreamSettings StreamSettings `json:"stream_settings"`
}

type Response struct {
	ChatUUID    uuid.UUID `json:"chat_uuid"`
	Answer      string    `json:"answer"`
	InputToken  int       `json:"input_token"`
	OutputToken int       `json:"output_token"`
}

type chatRequestContext struct {
	ctx            context.Context
	span           trace.Span
	req            QuestionRequest
	chat           *goai.ChatHistory
	messages       []goai.LLMMessage
	llmCompletion  *goai.LLMRequest
	logger         *log.Logger
	historyStorage goai.ChatHistoryStorage
}

func prepareRequestContext(
	r *http.Request,
	logger *log.Logger,
	historyStorage goai.ChatHistoryStorage,
	toolsProvider *goai.ToolsProvider,
	mcpClient *mcp.Client,
	operationName string,
) (*chatRequestContext, error) {
	ctx, span := observability.StartSpan(r.Context(), operationName)

	var req QuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	// Validate request
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	// Add observability attributes
	addRequestAttributes(ctx, req)

	// Initialize chat and get history
	chat, messages, err := initializeChatAndHistory(ctx, req, historyStorage, logger, mcpClient)
	if err != nil {
		return nil, err
	}

	// Add user message
	messages, err = addUserMessage(ctx, messages, req.Question, chat.UUID, historyStorage)
	if err != nil {
		return nil, err
	}

	// Setup LLM
	llmCompletion, err := setupLLMCompletion(ctx, req, toolsProvider)
	if err != nil {
		return nil, err
	}

	return &chatRequestContext{
		ctx:            ctx,
		span:           span,
		req:            req,
		chat:           chat,
		messages:       messages,
		llmCompletion:  llmCompletion,
		logger:         logger,
		historyStorage: historyStorage,
	}, nil
}

func HandleAsk(mcpClient *mcp.Client, logger *log.Logger, historyStorage goai.ChatHistoryStorage, toolsProvider *goai.ToolsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx, err := prepareRequestContext(r, logger, historyStorage, toolsProvider, mcpClient, "handle_ask")
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err.Error(), err, r.Context())
			return
		}
		defer reqCtx.span.End()

		// Generate response
		generateCtx, generateSpan := observability.StartSpan(reqCtx.ctx, "generate_response")
		defer generateSpan.End()

		response, err := generateSynchronousResponse(generateCtx, reqCtx)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error(), err, generateCtx)
			return
		}

		// Return the successful response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			ChatUUID:    reqCtx.chat.UUID,
			Answer:      response.Text,
			InputToken:  response.TotalInputToken,
			OutputToken: response.TotalOutputToken,
		})
	}
}

func HandleAskStream(mcpClient *mcp.Client, logger *log.Logger, historyStorage goai.ChatHistoryStorage, toolsProvider *goai.ToolsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx, err := prepareRequestContext(r, logger, historyStorage, toolsProvider, mcpClient, "handle_ask_stream")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer reqCtx.span.End()

		// Setup streaming
		if err := setupStreamingHeaders(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		err = handleStreamingResponse(reqCtx, w, flusher)
		if err != nil {
			logger.Printf("Streaming error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Helper functions

func validateRequest(req QuestionRequest) error {
	if req.Question == "" {
		return errors.New("question cannot be empty")
	}
	if req.LLMProvider.Provider == "" || req.LLMProvider.ModelID == "" {
		return errors.New("LLM provider is required")
	}
	supportedLLMProviders := getLLMProviders()
	if !supportedLLMProviders.IsSupported(req.LLMProvider.Provider, req.LLMProvider.ModelID) {
		return errors.New("LLM provider or model is not supported")
	}
	return nil
}

func setupLLMCompletion(ctx context.Context, req QuestionRequest, toolsProvider *goai.ToolsProvider) (*goai.LLMRequest, error) {
	reqOptions := prepareLLMRequestOptions(req)
	if len(req.SelectedTools) > 0 {
		reqOptions = append(reqOptions,
			goai.UseToolsProvider(toolsProvider),
			goai.WithAllowedTools(req.SelectedTools),
		)
	}

	builder := llm.NewLLMBuilder(ctx)
	llmProvider, err := builder.BuildProvider(llm.ProviderConfig{
		Provider: req.LLMProvider.Provider,
		ModelID:  req.LLMProvider.ModelID,
	})
	if err != nil {
		return nil, err
	}

	return goai.NewLLMRequest(goai.NewRequestConfig(reqOptions...), llmProvider), nil
}

func handleStreamingResponse(reqCtx *chatRequestContext, w http.ResponseWriter, flusher http.Flusher) error {
	streamChan, err := reqCtx.llmCompletion.GenerateStream(reqCtx.ctx, reqCtx.messages)
	if err != nil {
		return err
	}

	var fullResponse strings.Builder
	for streamResp := range streamChan {
		if streamResp.Error != nil {
			return streamResp.Error
		}

		if err := writeStreamChunk(w, flusher, streamResp); err != nil {
			return err
		}

		fullResponse.WriteString(streamResp.Text)

		if streamResp.Done {
			return saveAssistantResponse(reqCtx, fullResponse.String())
		}
	}
	return nil
}

func addRequestAttributes(ctx context.Context, req QuestionRequest) {
	observability.AddAttribute(ctx, "question.length", len(req.Question))
	observability.AddAttribute(ctx, "question.use_tools", req.SelectedTools)
	observability.AddAttribute(ctx, "model.temperature", req.ModelSettings.Temperature)
	observability.AddAttribute(ctx, "model.max_tokens", req.ModelSettings.MaxTokens)
	observability.AddAttribute(ctx, "model.top_p", req.ModelSettings.TopP)
	observability.AddAttribute(ctx, "model.top_k", req.ModelSettings.TopK)
	observability.AddAttribute(ctx, "llm.provider", req.LLMProvider.Provider)
	observability.AddAttribute(ctx, "llm.model_id", req.LLMProvider.ModelID)

	if len(req.SelectedTools) > 0 {
		observability.AddAttribute(ctx, "tools.enabled", true)
		observability.ToolsEnabledTotal.WithLabelValues(req.LLMProvider.Provider, req.LLMProvider.ModelID).Inc()

		for _, tool := range req.SelectedTools {
			observability.ToolsUsageTotal.WithLabelValues(
				tool,
				req.LLMProvider.Provider,
				req.LLMProvider.ModelID,
			).Inc()
		}
	}
}

func initializeChatAndHistory(
	ctx context.Context,
	req QuestionRequest,
	historyStorage goai.ChatHistoryStorage,
	logger *log.Logger,
	mcpClient *mcp.Client,
) (*goai.ChatHistory, []goai.LLMMessage, error) {
	chat, err := getOrInitializeChat(req, historyStorage, logger, ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get or create chat history: %w", err)
	}

	messages, err := getTruncatedChatHistory(ctx, chat.UUID, historyStorage)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve chat history: %w", err)
	}

	if len(messages) == 0 {
		promptMessages, err := buildMessagesFromPromptTemplates(ctx, mcpClient, req)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build prompt templates: %w", err)
		}
		messages = append(messages, promptMessages...)
	}

	return chat, messages, nil
}

func addUserMessage(
	ctx context.Context,
	messages []goai.LLMMessage,
	question string,
	chatUUID uuid.UUID,
	historyStorage goai.ChatHistoryStorage,
) ([]goai.LLMMessage, error) {
	userMessage := goai.LLMMessage{
		Role: goai.UserRole,
		Text: question,
	}
	messages = append(messages, userMessage)

	err := historyStorage.AddMessage(ctx, chatUUID, goai.ChatHistoryMessage{
		LLMMessage:  userMessage,
		GeneratedAt: time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add user message to chat history: %w", err)
	}

	return messages, nil
}

func setupStreamingHeaders(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	return nil
}

func writeStreamChunk(w http.ResponseWriter, flusher http.Flusher, streamResp goai.StreamingLLMResponse) error {
	response := struct {
		Content string `json:"content"`
		MetaKey string `json:"meta_key,omitempty"`
		Done    bool   `json:"done,omitempty"`
	}{
		Content: streamResp.Text,
		Done:    streamResp.Done,
	}

	chunkData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal stream chunk: %w", err)
	}

	if _, err := fmt.Fprintf(w, "%s\n", chunkData); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	flusher.Flush()
	return nil
}

func saveAssistantResponse(reqCtx *chatRequestContext, response string) error {
	err := reqCtx.historyStorage.AddMessage(reqCtx.ctx, reqCtx.chat.UUID, goai.ChatHistoryMessage{
		LLMMessage: goai.LLMMessage{
			Role: goai.AssistantRole,
			Text: response,
		},
		GeneratedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to add assistant message to history: %w", err)
	}
	return nil
}

func generateSynchronousResponse(ctx context.Context, reqCtx *chatRequestContext) (*goai.LLMResponse, error) {
	observability.AddAttribute(ctx, "HandleAsk.total_messages", len(reqCtx.messages))

	timer := prometheus.NewTimer(observability.LLMCompletionDuration.WithLabelValues(
		reqCtx.req.LLMProvider.Provider,
		reqCtx.req.LLMProvider.ModelID,
		"success",
	))
	defer timer.ObserveDuration()

	// Track in-flight requests
	inFlightMetric := observability.LLMCompletionInFlight.WithLabelValues(
		reqCtx.req.LLMProvider.Provider,
		reqCtx.req.LLMProvider.ModelID,
	)
	inFlightMetric.Inc()
	defer inFlightMetric.Dec()

	response, err := reqCtx.llmCompletion.Generate(ctx, reqCtx.messages)
	if err != nil {
		// Record failed completion duration with "error" status
		observability.LLMCompletionDuration.WithLabelValues(
			reqCtx.req.LLMProvider.Provider,
			reqCtx.req.LLMProvider.ModelID,
			"error",
		).Observe(timer.ObserveDuration().Seconds())

		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Update metrics
	observability.TokensInputTotal.WithLabelValues(
		reqCtx.req.LLMProvider.Provider,
		reqCtx.req.LLMProvider.ModelID,
	).Add(float64(response.TotalInputToken))

	observability.TokensOutputTotal.WithLabelValues(
		reqCtx.req.LLMProvider.Provider,
		reqCtx.req.LLMProvider.ModelID,
	).Add(float64(response.TotalOutputToken))

	// Add response to chat history
	err = saveAssistantResponse(reqCtx, response.Text)
	if err != nil {
		return nil, err
	}

	// Add attributes for observability
	observability.AddAttribute(ctx, "response.input_tokens", response.TotalInputToken)
	observability.AddAttribute(ctx, "response.output_tokens", response.TotalOutputToken)

	return &response, nil
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

func writeErrorResponse(w http.ResponseWriter, status int, message string, err error, ctx context.Context) {
	if err != nil {
		observability.AddAttribute(ctx, "error", err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
