// service/llm/builder.go
package llm

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/shaharia-lab/goai"
	"os"
	"strings"
)

type ProviderConfig struct {
	Provider string
	ModelID  string
}

type LLMBuilder struct {
	ctx context.Context
}

func NewLLMBuilder(ctx context.Context) *LLMBuilder {
	return &LLMBuilder{ctx: ctx}
}

func (b *LLMBuilder) BuildProvider(config ProviderConfig) (goai.LLMProvider, error) {
	switch strings.ToLower(config.Provider) {
	case "anthropic":
		return b.buildAnthropicProvider(config.ModelID)
	case "openai":
		return b.buildOpenAIProvider(config.ModelID)
	case "amazon bedrock":
		return b.buildBedrockProvider(config.ModelID)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}
}

func (b *LLMBuilder) buildAnthropicProvider(modelID string) (goai.LLMProvider, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}

	return goai.NewAnthropicLLMProvider(goai.AnthropicProviderConfig{
		Client: goai.NewAnthropicClient(apiKey),
		Model:  modelID,
	}), nil
}

func (b *LLMBuilder) buildOpenAIProvider(modelID string) (goai.LLMProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	return goai.NewOpenAILLMProvider(goai.OpenAIProviderConfig{
		Client: goai.NewOpenAIClient(apiKey),
		Model:  modelID,
	}), nil
}

func (b *LLMBuilder) buildBedrockProvider(modelID string) (goai.LLMProvider, error) {
	apiKey := os.Getenv("AMAZON_BEDROCK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("AMAZON_BEDROCK_API_KEY is required")
	}

	awsConfig, err := config.LoadDefaultConfig(b.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return goai.NewBedrockLLMProvider(goai.BedrockProviderConfig{
		Client: bedrockruntime.NewFromConfig(awsConfig),
		Model:  modelID,
	}), nil
}
