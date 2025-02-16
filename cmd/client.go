package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/mcp-kit/pkg/config"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"time"
)

func NewTaskCmd(logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "task",
		Short: "Ask a question or give a task to the LLM model",
		Long:  "Ask a question or give a task to the LLM model",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			ctx := context.Background()

			sseClient, err := initializeSSEClient(cfg, logger)
			if err != nil {
				return err
			}
			defer sseClient.Close(ctx)

			llm, err := initializeLLM(sseClient)
			if err != nil {
				return err
			}

			input, err := readUserInput()
			if err != nil {
				return err
			}

			return generateResponse(ctx, llm, input)
		},
	}
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func initializeSSEClient(cfg *config.Config, logger *log.Logger) (*mcp.Client, error) {
	client := mcp.NewClient(mcp.NewSSETransport(), mcp.ClientConfig{
		ClientName:    "My MCP Kit Client",
		ClientVersion: "1.0.0",
		Logger:        log.New(logger.Writer(), "", log.LstdFlags),
		RetryDelay:    5 * time.Second,
		MaxRetries:    3,
		SSE: mcp.SSEConfig{
			URL: cfg.MCPServerURL,
		},
	})

	if err := client.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("SSE Client failed to connect: %w", err)
	}

	return client, nil
}

func initializeLLM(sseClient *mcp.Client) (*goai.LLMRequest, error) {
	toolsProvider := goai.NewToolsProvider()
	if err := toolsProvider.AddMCPClient(sseClient); err != nil {
		return nil, fmt.Errorf("failed to add MCP client: %w", err)
	}

	llmProvider := goai.NewOpenAILLMProvider(goai.OpenAIProviderConfig{
		Client: goai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY")),
		Model:  openai.ChatModelGPT3_5Turbo,
	})

	llm := goai.NewLLMRequest(goai.NewRequestConfig(
		goai.WithMaxToken(100),
		goai.WithTemperature(0.7),
		goai.UseToolsProvider(toolsProvider),
	), llmProvider)

	return llm, nil
}

func readUserInput() (string, error) {
	fmt.Println("Enter your text (press Enter when done):")
	scanner := bufio.NewScanner(os.Stdin)

	var input string
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			break
		}
		input += text + "\n"
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}

	return strings.TrimSpace(input), nil
}

func generateResponse(ctx context.Context, llm *goai.LLMRequest, input string) error {
	response, err := llm.Generate(ctx, []goai.LLMMessage{
		{Role: goai.UserRole, Text: input},
	})
	if err != nil {
		return fmt.Errorf("failed to generate response: %w", err)
	}

	fmt.Printf("Response: %s\n", response.Text)
	fmt.Printf("Input token: %d, Output token: %d\n", response.TotalInputToken, response.TotalOutputToken)

	return nil
}
