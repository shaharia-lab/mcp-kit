package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/fatih/color"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/mcp-kit/internal/config"
	"github.com/shaharia-lab/mcp-kit/internal/tools"
	"github.com/spf13/cobra"
)

func NewTaskCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "task",
		Short: "Ask a question or give a task to the LLM model",
		Long:  "Ask a question or give a task to the LLM model",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			llm, err := initializeLLM()
			if err != nil {
				printError("Failed to initialize LLM", err)
				return err
			}

			input, err := readUserInput()
			if err != nil {
				printError("Failed to read user input", err)
				return err
			}

			return generateResponse(ctx, llm, input)
		},
	}
}

func loadConfig(configFilePath string) (*config.Config, error) {
	cfg, err := config.Load(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func initializeLLM() (*goai.LLMRequest, error) {
	toolsProvider := goai.NewToolsProvider()
	/*if err := toolsProvider.AddMCPClient(sseClient); err != nil {
		return nil, fmt.Errorf("failed to add MCP client: %w", err)
	}*/
	toolsProvider.AddTools(tools.MCPToolsRegistry)

	/*llmProvider := goai.NewOpenAILLMProvider(goai.OpenAIProviderConfig{
		Client: goai.NewOpenAIClient(os.Getenv("OPENAI_API_KEY")),
		Model:  openai.ChatModelGPT3_5Turbo,
	})*/

	llmProvider := goai.NewAnthropicLLMProvider(goai.AnthropicProviderConfig{
		Client: goai.NewAnthropicClient(os.Getenv("ANTHROPIC_API_KEY")),
		Model:  anthropic.ModelClaude3_5Sonnet20241022,
	})

	llm := goai.NewLLMRequest(goai.NewRequestConfig(
		goai.WithMaxToken(1000),
		goai.WithTemperature(0.5),
		goai.UseToolsProvider(toolsProvider),
	), llmProvider)

	return llm, nil
}

func readUserInput() (string, error) {
	color.New(color.FgGreen).Println("Enter your text (press Enter when done):")
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

	color.New(color.FgYellow).Printf("Your input: %s\n", strings.TrimSpace(input))
	return strings.TrimSpace(input), nil
}

func generateResponse(ctx context.Context, llm *goai.LLMRequest, input string) error {
	response, err := llm.Generate(ctx, []goai.LLMMessage{
		{Role: goai.UserRole, Text: fmt.Sprintf("%s", input)},
	})
	if err != nil {
		printError("Failed to generate response", err)
		return err
	}

	color.New(color.FgBlue).Println("---------- Response ----------")
	color.New(color.FgBlue).Printf("%s\n", response.Text)
	color.New(color.FgBlue).Println("-----------------------------")

	color.New(color.FgMagenta).Printf("Input token: %d, Output token: %d\n", response.TotalInputToken, response.TotalOutputToken)

	return nil
}

func printError(msg string, err error) {
	color.New(color.FgRed, color.Bold).Printf("[ERROR] %s: %v\n", msg, err)
}
