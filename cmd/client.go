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

			// start buffer
			// Create a scanner to read from standard input
			fmt.Println("Enter your text (press Enter when done):")
			scanner := bufio.NewScanner(os.Stdin)

			// Collect all input lines
			var input string
			for scanner.Scan() {
				text := scanner.Text()
				// Break if user enters an empty line
				if text == "" {
					break
				}
				input += text + "\n"
			}

			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}

			// Trim the trailing newline
			input = strings.TrimSpace(input)

			// Generate response
			response, err := llm.Generate(ctx, []goai.LLMMessage{
				{Role: goai.UserRole, Text: input},
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
