package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shaharia-lab/goai/mcp"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// FileInfo represents file metadata
type FileInfo struct {
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	AccessedAt  time.Time `json:"accessed_at"`
	IsDirectory bool      `json:"is_directory"`
	Permissions string    `json:"permissions"`
}

var FilesystemTools = []mcp.Tool{
	{
		Name:        "list_directory",
		Description: "List directory contents with [FILE] or [DIR] prefixes",
		InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "path": {
                "type": "string",
                "description": "Path to the directory to list"
            }
        },
        "required": ["path"]
    }`),
		Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
			var input struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(params.Arguments, &input); err != nil {
				return mcp.CallToolResult{}, err
			}

			entries, err := os.ReadDir(input.Path)
			if err != nil {
				return mcp.CallToolResult{}, err
			}

			var listing strings.Builder
			for _, entry := range entries {
				prefix := "[FILE]"
				if entry.IsDir() {
					prefix = "[DIR]"
				}
				fmt.Fprintf(&listing, "%s %s\n", prefix, entry.Name())
			}

			return mcp.CallToolResult{
				Content: []mcp.ToolResultContent{
					{
						Type: "text",
						Text: listing.String(),
					},
				},
			}, nil
		},
	},
	{
		Name:        "read_file",
		Description: "Read complete contents of a file with UTF-8 encoding",
		InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "path": {
                    "type": "string",
                    "description": "Path to the file to read"
                }
            },
            "required": ["path"]
        }`),
		Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
			var input struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(params.Arguments, &input); err != nil {
				return mcp.CallToolResult{}, err
			}

			content, err := ioutil.ReadFile(input.Path)
			if err != nil {
				return mcp.CallToolResult{}, err
			}

			return mcp.CallToolResult{
				Content: []mcp.ToolResultContent{
					{
						Type: "text",
						Text: string(content),
					},
				},
			}, nil
		},
	},
	{
		Name:        "write_file",
		Description: "Create new file or overwrite existing file with content",
		InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "path": {
                    "type": "string",
                    "description": "Path where to write the file"
                },
                "content": {
                    "type": "string",
                    "description": "Content to write to the file"
                }
            },
            "required": ["path", "content"]
        }`),
		Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
			var input struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(params.Arguments, &input); err != nil {
				return mcp.CallToolResult{}, err
			}

			if err := ioutil.WriteFile(input.Path, []byte(input.Content), 0644); err != nil {
				return mcp.CallToolResult{}, err
			}

			return mcp.CallToolResult{
				Content: []mcp.ToolResultContent{
					{
						Type: "text",
						Text: fmt.Sprintf("Successfully wrote %d bytes to %s", len(input.Content), input.Path),
					},
				},
			}, nil
		},
	},
	{
		Name:        "get_file_info",
		Description: "Get detailed file/directory metadata",
		InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "path": {
                    "type": "string",
                    "description": "Path to the file or directory"
                }
            },
            "required": ["path"]
        }`),
		Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
			var input struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(params.Arguments, &input); err != nil {
				return mcp.CallToolResult{}, err
			}

			info, err := os.Stat(input.Path)
			if err != nil {
				return mcp.CallToolResult{}, err
			}

			fileInfo := FileInfo{
				Size:        info.Size(),
				ModifiedAt:  info.ModTime(),
				IsDirectory: info.IsDir(),
				Permissions: info.Mode().String(),
			}

			infoJSON, err := json.MarshalIndent(fileInfo, "", "  ")
			if err != nil {
				return mcp.CallToolResult{}, err
			}

			return mcp.CallToolResult{
				Content: []mcp.ToolResultContent{
					{
						Type: "text",
						Text: string(infoJSON),
					},
				},
			}, nil
		},
	},
}
