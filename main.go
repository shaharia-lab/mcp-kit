// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shaharia-lab/goai/mcp"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP (Model Context Protocol) server and client",
	Long:  `MCP (Model Context Protocol) server and client`,
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Long:  `Start the server with specified configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create base server
		baseServer, err := mcp.NewBaseServer(
			mcp.UseLogger(log.New(os.Stderr, "[MCP] ", log.LstdFlags)),
		)
		if err != nil {
			panic(err)
		}

		// Add tool
		tool := mcp.Tool{
			Name:        "greet",
			Description: "Greet user",
			InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "name": {"type": "string"}
            },
            "required": ["name"]
        }`),
			Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
				var input struct {
					Name string `json:"name"`
				}
				json.Unmarshal(params.Arguments, &input)
				return mcp.CallToolResult{
					Content: []mcp.ToolResultContent{{
						Type: "text",
						Text: fmt.Sprintf("Hello, %s!", input.Name),
					}},
				}, nil
			},
		}
		baseServer.AddTools(tool)

		// Create and run SSE server
		server := mcp.NewSSEServer(baseServer)
		server.SetAddress(":8080")

		ctx := context.Background()
		if err := server.Run(ctx); err != nil {
			return err
		}

		return nil
	},
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start the client",
	Long:  `Start the client and connect to the server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sseConfig := mcp.ClientConfig{
			ClientName:    "MySSEClient",
			ClientVersion: "1.0.0",
			Logger:        log.New(os.Stdout, "[SSE] ", log.LstdFlags),
			RetryDelay:    5 * time.Second,
			MaxRetries:    3,
			SSE: mcp.SSEConfig{
				URL: "http://localhost:8080/events", // Replace with your SSE endpoint
			},
		}

		sseTransport := mcp.NewSSETransport()
		sseClient := mcp.NewClient(sseTransport, sseConfig)

		if err := sseClient.Connect(); err != nil {
			log.Fatalf("SSE Client failed to connect: %v", err)
		}
		defer sseClient.Close()

		tools, err := sseClient.ListTools()
		if err != nil {
			return fmt.Errorf("failed to list tools (SSE): %w", err)
		}
		fmt.Printf("SSE Tools: %+v\n", tools)

		return nil
	},
}

func init() {
	// Server flags
	serverCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")

	// Client flags
	clientCmd.Flags().StringP("server", "s", "localhost:8080", "Server address to connect to")

	// Add commands to root command
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(clientCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
