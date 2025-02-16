package cmd

import (
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/mcp-kit/pkg/config"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

func NewClientCmd(logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "get_weather",
		Short: "Get the current weather for a given location",
		Long:  "Get the current weather for a given location",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			ctx := context.Background()

			sseClient := mcp.NewClient(mcp.NewSSETransport(), mcp.ClientConfig{
				ClientName:    "My MCP Kit Client",
				ClientVersion: "1.0.0",
				Logger:        log.New(logger.Writer(), "", log.LstdFlags),
				RetryDelay:    5 * time.Second,
				MaxRetries:    3,
				SSE: mcp.SSEConfig{
					URL: cfg.MCPServerURL,
				},
			})

			toolsProvider := goai.NewToolsProvider()
			err = toolsProvider.AddMCPClient(sseClient)
			if err != nil {
				return fmt.Errorf("failed to add MCP client: %w", err)
			}

			llmProvider := goai.NewOpenAILLMProvider(goai.OpenAIProviderConfig{
				Client: goai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY")),
				Model:  openai.ChatModelGPT3_5Turbo,
			})

			// Configure LLM Request
			llm := goai.NewLLMRequest(goai.NewRequestConfig(
				goai.WithMaxToken(100),
				goai.WithTemperature(0.7),
				goai.UseToolsProvider(toolsProvider),
			), llmProvider)

			// let's connect SSE Client
			if err := sseClient.Connect(ctx); err != nil {
				log.Fatalf("SSE Client failed to connect: %v", err)
			}
			defer sseClient.Close(ctx)

			// Generate response
			response, err := llm.Generate(ctx, []goai.LLMMessage{
				{Role: goai.UserRole, Text: "What's the weather in Dhaka?"},
			})

			if err != nil {
				panic(err)
			}

			fmt.Printf("Response: %s\n", response.Text)
			fmt.Printf("Input token: %d, Output token: %d", response.TotalInputToken, response.TotalOutputToken)

			return nil
		},
	}
}
