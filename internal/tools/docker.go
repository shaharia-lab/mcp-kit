package tools

import (
	"context"
	"encoding/json"
	"os/exec"

	"github.com/shaharia-lab/goai/mcp"
)

var dockerAllInOneTool = mcp.Tool{
	Name:        "docker_all_in_one",
	Description: "Execute any Docker command",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "command": {
                "type": "string",
                "description": "Docker command to execute"
            },
            "args": {
                "type": "array",
                "items": {
                    "type": "string"
                },
                "description": "Arguments for the Docker command"
            }
        },
        "required": ["command"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := append([]string{input.Command}, input.Args...)
		cmd := exec.CommandContext(ctx, "docker", args...)
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
