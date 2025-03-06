package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go"
)

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

func (s SupportedLLMProviders) IsSupported(providerName string, modelID string) bool {
	for _, provider := range s.Providers {
		if provider.Name == providerName {
			for _, model := range provider.Models {
				if model.ModelID == modelID {
					return true
				}
			}
		}
	}
	return false
}

// getLLMProviders retrieves a list of supported LLM providers.
func getLLMProviders() SupportedLLMProviders {
	return SupportedLLMProviders{
		Providers: []Provider{
			{
				Name: "Anthropic",
				Models: []Model{
					{
						Name:        "Claude 3.5 Haiku Latest",
						Description: "Fast and cost-effective model",
						ModelID:     anthropic.ModelClaude3_5HaikuLatest,
					},
					{
						Name:        "Claude 3.5 Haiku 2024-10-22",
						Description: "Fast and cost-effective model",
						ModelID:     anthropic.ModelClaude3_5Haiku20241022,
					},
					{
						Name:        "Claude 3.7 Sonnet",
						Description: "Most intelligent model from Anthropic",
						ModelID:     anthropic.ModelClaude3_7SonnetLatest,
					},
					{
						Name:        "Claude 3.5 Sonnet Latest",
						Description: "Our most intelligent model",
						ModelID:     anthropic.ModelClaude3_5SonnetLatest,
					},
					{
						Name:        "Claude 3.5 Sonnet 2024-10-22",
						Description: "Our most intelligent model",
						ModelID:     anthropic.ModelClaude3_5Sonnet20241022,
					},
					{
						Name:        "Claude 3.5 Sonnet 2024-06-20",
						Description: "Our previous most intelligent model",
						ModelID:     anthropic.ModelClaude_3_5_Sonnet_20240620,
					},
					{
						Name:        "Claude 3 Opus Latest",
						Description: "Excels at writing and complex tasks",
						ModelID:     anthropic.ModelClaude3OpusLatest,
					},
					{
						Name:        "Claude 3 Opus 2024-02-29",
						Description: "Excels at writing and complex tasks",
						ModelID:     anthropic.ModelClaude_3_Opus_20240229,
					},
					{
						Name:        "Claude 3 Sonnet 2024-02-29",
						Description: "Balance of speed and intelligence",
						ModelID:     anthropic.ModelClaude_3_Sonnet_20240229,
					},
					{
						Name:        "Claude 3 Haiku 2024-03-07",
						Description: "Our previous fast and cost-effective",
						ModelID:     anthropic.ModelClaude_3_Haiku_20240307,
					},
					{
						Name:        "Claude 2.1",
						Description: "Powerful language model for general-purpose tasks",
						ModelID:     anthropic.ModelClaude_2_1,
					},
					{
						Name:        "Claude 2.0",
						Description: "Advanced language model optimized for reliability and thoughtful responses",
						ModelID:     anthropic.ModelClaude_2_0,
					},
				},
			},
			{
				Name: "OpenAI",
				Models: []Model{
					{
						Name:        "GPT-4o Latest",
						Description: "Latest GPT-4o model",
						ModelID:     openai.ChatModelChatgpt4oLatest,
					},
					{
						Name:        "GPT-4o Mini",
						Description: "Optimized GPT-4o Mini model",
						ModelID:     openai.ChatModelGPT4oMini,
					},
					{
						Name:        "GPT-4",
						Description: "Standard GPT-4 model",
						ModelID:     openai.ChatModelGPT4,
					},
					{
						Name:        "GPT-4 Turbo",
						Description: "Most capable GPT-4 model for various tasks",
						ModelID:     openai.ChatModelGPT4Turbo,
					},
					{
						Name:        "GPT-3.5 Turbo",
						Description: "Efficient model balancing performance and speed",
						ModelID:     openai.ChatModelGPT3_5Turbo,
					},
					{
						Name:        "GPT-4.5 Preview",
						Description: "Last GPT-4.5 model from OpenAI",
						ModelID:     openai.ChatModelGPT4_5Preview,
					},
				},
			},
			{
				Name: "Amazon Bedrock",
				Models: []Model{
					{
						Name:        "Claude 3 Haiku 2024-03-07",
						Description: "Optimized for quick, detailed responses",
						ModelID:     "anthropic.claude-3-haiku-20240307-v1:0",
					},
					{
						Name:        "Claude 3 Opus 2024-02-29",
						Description: "Excels at writing and complex tasks",
						ModelID:     "anthropic.claude-3-opus-20240229-v1:0",
					},
					{
						Name:        "Claude 3 Sonnet 2024-02-29",
						Description: "Balanced performance and intelligence",
						ModelID:     "anthropic.claude-3-sonnet-20240229-v1:0",
					},
					{
						Name:        "Claude 3.5 Haiku 2024-10-22",
						Description: "Our most recent fast and cost-effective model",
						ModelID:     "anthropic.claude-3-5-haiku-20241022-v1:0",
					},
					{
						Name:        "Claude 3.5 Sonnet 2024-10-22",
						Description: "Intelligent and fine-tuned for deep tasks",
						ModelID:     "anthropic.claude-3-5-sonnet-20241022-v2:0",
					},
					{
						Name:        "Claude 3.5 Sonnet 2024-06-20",
						Description: "Balanced for intelligent and previous updates",
						ModelID:     "anthropic.claude-3-5-sonnet-20240620-v1:0",
					},
					{
						Name:        "Claude 3.7 Sonnet",
						Description: "Latest best model from Anthropic",
						ModelID:     "anthropic.claude-3-7-sonnet-20250219-v1:0",
					},
					{
						Name:        "Titan Text G1 - Express",
						Description: "Amazon's express text model for versatile use cases",
						ModelID:     "amazon.titan-text-express-v1",
					},
					{
						Name:        "Cohere: Command R+",
						Description: "Advanced command response model",
						ModelID:     "cohere.command-r-plus-v1:0",
					},
					{
						Name:        "Cohere: Command R",
						Description: "Command-response optimized model",
						ModelID:     "cohere.command-r-v1:0",
					},
					{
						Name:        "Llama 3 8B Instruct",
						Description: "Meta's mid-range instruct model",
						ModelID:     "meta.llama3-8b-instruct-v1:0",
					},
					{
						Name:        "Llama 3 70B Instruct",
						Description: "Meta's large instruct model",
						ModelID:     "meta.llama3-70b-instruct-v1:0",
					},
					{
						Name:        "Llama 3.1 8B Instruct",
						Description: "Updated 8B instruct model by Meta",
						ModelID:     "meta.llama3-1-8b-instruct-v1:0",
					},
					{
						Name:        "Llama 3.1 70B Instruct",
						Description: "Updated comprehensive instruct model by Meta",
						ModelID:     "meta.llama3-1-70b-instruct-v1:0",
					},
					{
						Name:        "Llama 3.1 405B Instruct",
						Description: "Meta's groundbreaking large instruct model",
						ModelID:     "meta.llama3-1-405b-instruct-v1:0",
					},
					{
						Name:        "Llama 3.2 1B Instruct",
						Description: "Compact instruct model for lightweight tasks",
						ModelID:     "meta.llama3-2-1b-instruct-v1:0",
					},
					{
						Name:        "Llama 3.2 3B Instruct",
						Description: "Balanced model for intelligence and agility",
						ModelID:     "meta.llama3-2-3b-instruct-v1:0",
					},
					{
						Name:        "Llama 3.2 11B Instruct",
						Description: "High-precision instruct model at 11B scale",
						ModelID:     "meta.llama3-2-11b-instruct-v1:0",
					},
					{
						Name:        "Llama 3.2 90B Instruct",
						Description: "Meta's premier 90B-scale instruct model",
						ModelID:     "meta.llama3-2-90b-instruct-v1:0",
					},
					{
						Name:        "Llama 3.3 70B Instruct",
						Description: "Meta's latest iteration of 70B instruct",
						ModelID:     "meta.llama3-3-70b-instruct-v1:0",
					},
					{
						Name:        "Mistral 7B Instruct",
						Description: "Compact yet powerful instruct model by MistralAI",
						ModelID:     "mistral.mistral-7b-instruct-v0:2",
					},
					{
						Name:        "Mistral Large (24.02)",
						Description: "Latest large model optimized by MistralAI",
						ModelID:     "mistral.mistral-large-2402-v1:0",
					},
				},
			},
			{
				Name: "DeepSeek",
				Models: []Model{
					{
						Name:        "DeepSeek Chat",
						Description: "Conversational AI model optimized for interactive chats",
						ModelID:     "deepseek-chat",
					},
					{
						Name:        "DeepSeek Reasoner",
						Description: "Advanced reasoning model for analytical tasks",
						ModelID:     "deepseek-reasoner",
					},
				},
			},
		},
	}
}

// LLMProvidersHandler handles requests for fetching supported LLM providers.
func LLMProvidersHandler(w http.ResponseWriter, r *http.Request) {
	providers := getLLMProviders()

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode the response
	if err := json.NewEncoder(w).Encode(providers); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
