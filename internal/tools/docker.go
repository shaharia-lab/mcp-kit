package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/shaharia-lab/goai/mcp"
)

// listContainersTool lists running Docker containers
var listContainersTool = mcp.Tool{
	Name:        "docker_list_containers",
	Description: "List all running Docker containers on the local machine",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "all": {
                "type": "boolean",
                "description": "If true, show all containers (default shows just running)"
            }
        }
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			All bool `json:"all"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"ps"}
		if input.All {
			args = append(args, "-a")
		}

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

// containerLogsTool gets logs from a specific container
var containerLogsTool = mcp.Tool{
	Name:        "docker_get_container_logs",
	Description: "Get the logs from a specific Docker container",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "container_id": {
                "type": "string",
                "description": "Container ID or name"
            },
            "tail": {
                "type": "integer",
                "description": "Number of lines to show from the end of the logs"
            }
        },
        "required": ["container_id"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			ContainerID string `json:"container_id"`
			Tail        int    `json:"tail"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"logs"}
		if input.Tail > 0 {
			args = append(args, fmt.Sprintf("--tail=%d", input.Tail))
		}
		args = append(args, input.ContainerID)

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

// listImagesTool lists Docker images
var listImagesTool = mcp.Tool{
	Name:        "docker_list_images",
	Description: "List all Docker images on the local machine",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {}
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		cmd := exec.CommandContext(ctx, "docker", "images")
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

// pruneImagesTool prunes Docker images
var pruneImagesTool = mcp.Tool{
	Name:        "docker_prune_images",
	Description: "Remove unused Docker images older than specified days",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "days": {
                "type": "integer",
                "description": "Remove images older than specified days"
            }
        },
        "required": ["days"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Days int `json:"days"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "docker", "image", "prune", "-a", "--force",
			fmt.Sprintf("--filter=until=%dh", input.Days*24))
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

// runContainerTool runs a Docker container
var runContainerTool = mcp.Tool{
	Name:        "docker_run_container",
	Description: "Run a Docker container from a specified image",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "image": {
                "type": "string",
                "description": "Docker image name to run"
            },
            "detach": {
                "type": "boolean",
                "description": "Run container in background"
            },
            "ports": {
                "type": "array",
                "items": {
                    "type": "string"
                },
                "description": "Port mappings (e.g. ['8080:80'])"
            }
        },
        "required": ["image"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Image  string   `json:"image"`
			Detach bool     `json:"detach"`
			Ports  []string `json:"ports"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"run"}
		if input.Detach {
			args = append(args, "-d")
		}
		for _, port := range input.Ports {
			args = append(args, "-p", port)
		}
		args = append(args, input.Image)

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

// pullImageTool pulls a Docker image
var pullImageTool = mcp.Tool{
	Name:        "docker_pull_image",
	Description: "Pull a Docker image from a registry",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "image": {
                "type": "string",
                "description": "Docker image name to pull"
            }
        },
        "required": ["image"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Image string `json:"image"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "docker", "pull", input.Image)
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

// pruneContainersTool prunes stopped containers
var pruneContainersTool = mcp.Tool{
	Name:        "docker_prune_containers",
	Description: "Remove all stopped containers and optionally remove containers by age",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "force": {
                "type": "boolean",
                "description": "Force removal without confirmation"
            },
            "until": {
                "type": "integer",
                "description": "Remove containers created before N hours (optional)"
            }
        }
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Force bool `json:"force"`
			Until int  `json:"until"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"container", "prune"}
		if input.Force {
			args = append(args, "--force")
		}
		if input.Until > 0 {
			args = append(args, fmt.Sprintf("--filter=until=%dh", input.Until))
		}

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

// dockerInspectTool provides detailed information about Docker objects
var dockerInspectTool = mcp.Tool{
	Name:        "docker_inspect",
	Description: "Inspect Docker containers, images, volumes, or networks",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "target": {
                "type": "string",
                "description": "ID or name of the Docker object to inspect"
            },
            "format": {
                "type": "string",
                "description": "Format the output using Go template (optional)"
            },
            "type": {
                "type": "string",
                "enum": ["container", "image", "volume", "network"],
                "description": "Type of Docker object to inspect"
            }
        },
        "required": ["target", "type"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Target string `json:"target"`
			Format string `json:"format"`
			Type   string `json:"type"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{input.Type, "inspect"}

		if input.Format != "" {
			args = append(args, fmt.Sprintf("--format=%s", input.Format))
		}

		args = append(args, input.Target)

		cmd := exec.CommandContext(ctx, "docker", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, err
		}

		// Try to pretty-print JSON output if no specific format was requested
		if input.Format == "" {
			var prettyJSON map[string]interface{}
			if err := json.Unmarshal(output, &prettyJSON); err == nil {
				prettyOutput, err := json.MarshalIndent(prettyJSON, "", "    ")
				if err == nil {
					output = prettyOutput
				}
			}
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
