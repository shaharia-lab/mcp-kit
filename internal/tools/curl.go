package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shaharia-lab/goai/mcp"
	"os/exec"
	"strings"
)

// httpGetTool performs HTTP GET requests
var httpGetTool = mcp.Tool{
	Name:        "curl_get",
	Description: "Perform HTTP GET request to a specified URL with optional headers",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "url": {
                "type": "string",
                "description": "Target URL for the GET request"
            },
            "headers": {
                "type": "object",
                "description": "HTTP headers to include in the request",
                "additionalProperties": {
                    "type": "string"
                }
            },
            "insecure": {
                "type": "boolean",
                "description": "Allow insecure server connections when using SSL"
            }
        },
        "required": ["url"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			URL      string            `json:"url"`
			Headers  map[string]string `json:"headers"`
			Insecure bool              `json:"insecure"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"-s"}
		if input.Insecure {
			args = append(args, "-k")
		}

		for key, value := range input.Headers {
			args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
		}

		args = append(args, input.URL)
		cmd := exec.CommandContext(ctx, "curl", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, err
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: string(output),
				},
			},
		}, nil
	},
}

// httpPostJsonTool performs HTTP POST requests with JSON data
var httpPostJsonTool = mcp.Tool{
	Name:        "curl_post_json",
	Description: "Perform HTTP POST request with JSON data",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "url": {
                "type": "string",
                "description": "Target URL for the POST request"
            },
            "data": {
                "type": "object",
                "description": "JSON data to send in the request body"
            },
            "headers": {
                "type": "object",
                "description": "Additional HTTP headers to include in the request",
                "additionalProperties": {
                    "type": "string"
                }
            },
            "insecure": {
                "type": "boolean",
                "description": "Allow insecure server connections when using SSL"
            }
        },
        "required": ["url", "data"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			URL      string            `json:"url"`
			Data     json.RawMessage   `json:"data"`
			Headers  map[string]string `json:"headers"`
			Insecure bool              `json:"insecure"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"-s", "-X", "POST"}
		if input.Insecure {
			args = append(args, "-k")
		}

		// Add Content-Type header if not specified
		if input.Headers == nil {
			input.Headers = make(map[string]string)
		}
		if _, exists := input.Headers["Content-Type"]; !exists {
			input.Headers["Content-Type"] = "application/json"
		}

		for key, value := range input.Headers {
			args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
		}

		args = append(args, "-d", string(input.Data))
		args = append(args, input.URL)

		cmd := exec.CommandContext(ctx, "curl", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, err
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: string(output),
				},
			},
		}, nil
	},
}

// httpFormPostTool performs HTTP POST requests with form data
var httpFormPostTool = mcp.Tool{
	Name:        "curl_post_form",
	Description: "Perform HTTP POST request with form data",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "url": {
                "type": "string",
                "description": "Target URL for the POST request"
            },
            "form_data": {
                "type": "object",
                "description": "Form data key-value pairs",
                "additionalProperties": {
                    "type": "string"
                }
            },
            "headers": {
                "type": "object",
                "description": "Additional HTTP headers to include in the request",
                "additionalProperties": {
                    "type": "string"
                }
            },
            "insecure": {
                "type": "boolean",
                "description": "Allow insecure server connections when using SSL"
            }
        },
        "required": ["url", "form_data"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			URL      string            `json:"url"`
			FormData map[string]string `json:"form_data"`
			Headers  map[string]string `json:"headers"`
			Insecure bool              `json:"insecure"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"-s", "-X", "POST"}
		if input.Insecure {
			args = append(args, "-k")
		}

		for key, value := range input.Headers {
			args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
		}

		// Build form data
		formData := []string{}
		for key, value := range input.FormData {
			formData = append(formData, fmt.Sprintf("%s=%s", key, value))
		}
		args = append(args, "-d", strings.Join(formData, "&"))

		args = append(args, input.URL)

		cmd := exec.CommandContext(ctx, "curl", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, err
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: string(output),
				},
			},
		}, nil
	},
}

// httpCustomRequestTool performs custom HTTP requests
var httpCustomRequestTool = mcp.Tool{
	Name:        "curl_custom_request",
	Description: "Perform custom HTTP request with specified method and data",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "url": {
                "type": "string",
                "description": "Target URL for the request"
            },
            "method": {
                "type": "string",
                "description": "HTTP method (PUT, DELETE, PATCH, etc.)"
            },
            "data": {
                "type": "string",
                "description": "Data to send in the request body"
            },
            "headers": {
                "type": "object",
                "description": "HTTP headers to include in the request",
                "additionalProperties": {
                    "type": "string"
                }
            },
            "insecure": {
                "type": "boolean",
                "description": "Allow insecure server connections when using SSL"
            }
        },
        "required": ["url", "method"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			URL      string            `json:"url"`
			Method   string            `json:"method"`
			Data     string            `json:"data"`
			Headers  map[string]string `json:"headers"`
			Insecure bool              `json:"insecure"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"-s", "-X", strings.ToUpper(input.Method)}
		if input.Insecure {
			args = append(args, "-k")
		}

		for key, value := range input.Headers {
			args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
		}

		if input.Data != "" {
			args = append(args, "-d", input.Data)
		}

		args = append(args, input.URL)

		cmd := exec.CommandContext(ctx, "curl", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, err
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: string(output),
				},
			},
		}, nil
	},
}
