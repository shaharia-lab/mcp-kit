package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/shaharia-lab/goai/mcp"
)

// findContentInFilesTool searches for content in files using grep
var findContentInFilesTool = mcp.Tool{
	Name:        "search_file_contents",
	Description: "Search for specific content in files within a directory",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "directory": {
                "type": "string",
                "description": "Directory path to search in"
            },
            "pattern": {
                "type": "string",
                "description": "Text pattern to search for"
            },
            "file_pattern": {
                "type": "string",
                "description": "File pattern to filter (e.g., '*.go', '*.txt')"
            },
            "case_sensitive": {
                "type": "boolean",
                "description": "Whether to perform case-sensitive search"
            },
            "recursive": {
                "type": "boolean",
                "description": "Whether to search recursively in subdirectories"
            }
        },
        "required": ["directory", "pattern"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Directory     string `json:"directory"`
			Pattern       string `json:"pattern"`
			FilePattern   string `json:"file_pattern"`
			CaseSensitive bool   `json:"case_sensitive"`
			Recursive     bool   `json:"recursive"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{}
		if !input.CaseSensitive {
			args = append(args, "-i")
		}
		if input.Recursive {
			args = append(args, "-r")
		}
		args = append(args, "-n")             // Show line numbers
		args = append(args, "--color=always") // Colorize output
		args = append(args, input.Pattern)

		if input.FilePattern != "" {
			args = append(args, fmt.Sprintf("--include=%s", input.FilePattern))
		}

		args = append(args, input.Directory)

		cmd := exec.CommandContext(ctx, "grep", args...)
		output, err := cmd.CombinedOutput()
		if err != nil && !strings.Contains(string(output), "No such file or directory") {
			// Ignore grep's exit status when no matches found
			if exitError, ok := err.(*exec.ExitError); !ok || exitError.ExitCode() != 1 {
				return mcp.CallToolResult{}, err
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

// findLargeFilesTool finds files exceeding specified size
var findLargeFilesTool = mcp.Tool{
	Name:        "find_large_files",
	Description: "Find files larger than specified size in directory",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "directory": {
                "type": "string",
                "description": "Directory path to search in"
            },
            "min_size_mb": {
                "type": "number",
                "description": "Minimum file size in megabytes"
            },
            "file_pattern": {
                "type": "string",
                "description": "Optional file pattern to filter (e.g., '*.log')"
            }
        },
        "required": ["directory", "min_size_mb"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Directory   string  `json:"directory"`
			MinSizeMB   float64 `json:"min_size_mb"`
			FilePattern string  `json:"file_pattern"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		sizeInBytes := int64(input.MinSizeMB * 1024 * 1024)
		args := []string{input.Directory, "-type", "f", "-size", fmt.Sprintf("+%d", sizeInBytes)}

		if input.FilePattern != "" {
			args = append(args, "-name", input.FilePattern)
		}

		args = append(args, "-exec", "ls", "-lh", "{}", ";")

		cmd := exec.CommandContext(ctx, "find", args...)
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

// findDuplicateFilesTool finds duplicate files in directory
var findDuplicateFilesTool = mcp.Tool{
	Name:        "find_duplicate_files",
	Description: "Find duplicate files in directory based on content",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "directory": {
                "type": "string",
                "description": "Directory path to search in"
            },
            "file_pattern": {
                "type": "string",
                "description": "Optional file pattern to filter (e.g., '*.jpg')"
            },
            "min_size": {
                "type": "integer",
                "description": "Minimum file size in bytes to consider"
            }
        },
        "required": ["directory"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Directory   string `json:"directory"`
			FilePattern string `json:"file_pattern"`
			MinSize     int    `json:"min_size"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		// First find files and compute their hashes
		findArgs := []string{input.Directory, "-type", "f"}
		if input.FilePattern != "" {
			findArgs = append(findArgs, "-name", input.FilePattern)
		}
		if input.MinSize > 0 {
			findArgs = append(findArgs, "-size", fmt.Sprintf("+%dc", input.MinSize))
		}

		// Use find to get files, then pipe through sort and uniq to find duplicates
		cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf(
			`find %s -type f %s -exec md5sum {} \; | sort | uniq -w32 -dD`,
			input.Directory,
			strings.Join(findArgs[1:], " "),
		))

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

// findRecentlyModifiedTool finds recently modified files
var findRecentlyModifiedTool = mcp.Tool{
	Name:        "find_recently_modified",
	Description: "Find files modified within specified time period",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "directory": {
                "type": "string",
                "description": "Directory path to search in"
            },
            "minutes": {
                "type": "integer",
                "description": "Find files modified within last N minutes"
            },
            "file_pattern": {
                "type": "string",
                "description": "Optional file pattern to filter"
            }
        },
        "required": ["directory", "minutes"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Directory   string `json:"directory"`
			Minutes     int    `json:"minutes"`
			FilePattern string `json:"file_pattern"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{input.Directory, "-type", "f", "-mmin", fmt.Sprintf("-%d", input.Minutes)}

		if input.FilePattern != "" {
			args = append(args, "-name", input.FilePattern)
		}

		args = append(args, "-ls")

		cmd := exec.CommandContext(ctx, "find", args...)
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

// searchCodePatternTool searches for code patterns
var searchCodePatternTool = mcp.Tool{
	Name:        "search_code_pattern",
	Description: "Search for specific code patterns in source files",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "directory": {
                "type": "string",
                "description": "Directory path to search in"
            },
            "pattern": {
                "type": "string",
                "description": "Code pattern to search for"
            },
            "file_types": {
                "type": "array",
                "items": {
                    "type": "string"
                },
                "description": "File extensions to search (e.g., ['go', 'java', 'py'])"
            },
            "exclude_dirs": {
                "type": "array",
                "items": {
                    "type": "string"
                },
                "description": "Directories to exclude (e.g., ['vendor', 'node_modules'])"
            }
        },
        "required": ["directory", "pattern"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			Directory   string   `json:"directory"`
			Pattern     string   `json:"pattern"`
			FileTypes   []string `json:"file_types"`
			ExcludeDirs []string `json:"exclude_dirs"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"-n", "--color=always"}

		// Add file type patterns
		if len(input.FileTypes) > 0 {
			patterns := []string{}
			for _, ext := range input.FileTypes {
				patterns = append(patterns, fmt.Sprintf("--include=*.%s", ext))
			}
			args = append(args, patterns...)
		}

		// Add directory exclusions
		if len(input.ExcludeDirs) > 0 {
			for _, dir := range input.ExcludeDirs {
				args = append(args, fmt.Sprintf("--exclude-dir=%s", dir))
			}
		}

		args = append(args, "-r", input.Pattern, input.Directory)

		cmd := exec.CommandContext(ctx, "grep", args...)
		output, err := cmd.CombinedOutput()
		if err != nil && !strings.Contains(string(output), "No such file or directory") {
			if exitError, ok := err.(*exec.ExitError); !ok || exitError.ExitCode() != 1 {
				return mcp.CallToolResult{}, err
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
