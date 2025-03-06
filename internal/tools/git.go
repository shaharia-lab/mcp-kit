package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/goai/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var gitStatusTool = mcp.Tool{
	Name:        "git_status",
	Description: "Shows the working tree status",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            }
        },
        "required": ["repo_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		// Parse the input
		span.AddEvent("ParseInput")
		var input struct {
			RepoPath string `json:"repo_path"`
		}
		if err = json.Unmarshal(params.Arguments, &input); err != nil {
			span.SetAttributes(attribute.String("error_stage", "json_unmarshal"))
			return mcp.CallToolResult{}, err
		}

		span.SetAttributes(
			attribute.String("repo_path", input.RepoPath),
		)

		// Execute git command
		span.AddEvent("ExecuteGitCommand",
			trace.WithAttributes(
				attribute.String("command", "git"),
				attribute.StringSlice("args", []string{"-C", input.RepoPath, "status"}),
			),
		)

		cmdStart := time.Now()
		cmd := exec.CommandContext(ctx, "git", "-C", input.RepoPath, "status")
		output, err := cmd.CombinedOutput()
		cmdDuration := time.Since(cmdStart)

		span.SetAttributes(
			attribute.Float64("cmd_execution_time_ms", float64(cmdDuration.Milliseconds())),
		)

		if err != nil {
			span.SetAttributes(
				attribute.String("error_stage", "git_command"),
				attribute.String("cmd_output", string(output)),
				attribute.Int("exit_code", cmd.ProcessState.ExitCode()),
			)
			return mcp.CallToolResult{}, fmt.Errorf("git status error: %w", err)
		}

		// Success
		outputStr := string(output)
		outputLen := len(outputStr)

		span.SetAttributes(
			attribute.Int("output_length", outputLen),
			attribute.Bool("success", true),
		)

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: outputStr,
				},
			},
		}, nil
	},
}

var gitDiffTool = mcp.Tool{
	Name:        "git_diff",
	Description: "Shows differences between branches or commits",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            },
            "target": {
                "type": "string",
                "description": "Target branch or commit to compare with"
            }
        },
        "required": ["repo_path", "target"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath string `json:"repo_path"`
			Target   string `json:"target"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "git", "-C", input.RepoPath, "diff", input.Target)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git diff error: %w", err)
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

var gitCommitTool = mcp.Tool{
	Name:        "git_commit",
	Description: "Records changes to the repository",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            },
            "message": {
                "type": "string",
                "description": "Commit message"
            }
        },
        "required": ["repo_path", "message"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath string `json:"repo_path"`
			Message  string `json:"message"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "git", "-C", input.RepoPath, "commit", "-m", input.Message)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git commit error: %w", err)
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

var gitDiffUnstagedTool = mcp.Tool{
	Name:        "git_diff_unstaged",
	Description: "Shows changes in working directory not yet staged",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            }
        },
        "required": ["repo_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath string `json:"repo_path"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "git", "-C", input.RepoPath, "diff")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git diff unstaged error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitDiffStagedTool = mcp.Tool{
	Name:        "git_diff_staged",
	Description: "Shows changes that are staged for commit",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            }
        },
        "required": ["repo_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath string `json:"repo_path"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "git", "-C", input.RepoPath, "diff", "--cached")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git diff staged error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitAddTool = mcp.Tool{
	Name:        "git_add",
	Description: "Adds file contents to the staging area",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            },
            "files": {
                "type": "array",
                "items": {"type": "string"},
                "description": "Array of file paths to stage"
            }
        },
        "required": ["repo_path", "files"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath string   `json:"repo_path"`
			Files    []string `json:"files"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := append([]string{"-C", input.RepoPath, "add"}, input.Files...)
		cmd := exec.CommandContext(ctx, "git", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git add error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitResetTool = mcp.Tool{
	Name:        "git_reset",
	Description: "Unstages all staged changes",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            }
        },
        "required": ["repo_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath string `json:"repo_path"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "git", "-C", input.RepoPath, "reset")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git reset error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitLogTool = mcp.Tool{
	Name:        "git_log",
	Description: "Shows the commit logs",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            },
            "max_count": {
                "type": "number",
                "description": "Maximum number of commits to show"
            }
        },
        "required": ["repo_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath string `json:"repo_path"`
			MaxCount int    `json:"max_count"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git_log tool error: arguments unmarshal error: %w", err)
		}

		args := []string{"-C", input.RepoPath, "log", "--pretty=format:\"%H|%an|%ad|%s\""}
		if input.MaxCount > 0 {
			args = append(args, fmt.Sprintf("-n%d", input.MaxCount))
		}

		cmd := exec.CommandContext(ctx, "git", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git_log command execution error: git log error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitCreateBranchTool = mcp.Tool{
	Name:        "git_create_branch",
	Description: "Creates a new branch",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            },
            "branch_name": {
                "type": "string",
                "description": "Name of the new branch"
            },
            "start_point": {
                "type": "string",
                "description": "Starting point for the new branch"
            }
        },
        "required": ["repo_path", "branch_name"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath   string `json:"repo_path"`
			BranchName string `json:"branch_name"`
			StartPoint string `json:"start_point"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		args := []string{"-C", input.RepoPath, "branch", input.BranchName}
		if input.StartPoint != "" {
			args = append(args, input.StartPoint)
		}

		cmd := exec.CommandContext(ctx, "git", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git branch error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitCheckoutTool = mcp.Tool{
	Name:        "git_checkout",
	Description: "Switches branches",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            },
            "branch_name": {
                "type": "string",
                "description": "Name of branch to checkout"
            }
        },
        "required": ["repo_path", "branch_name"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath   string `json:"repo_path"`
			BranchName string `json:"branch_name"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "git", "-C", input.RepoPath, "checkout", input.BranchName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git checkout error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitShowTool = mcp.Tool{
	Name:        "git_show",
	Description: "Shows the contents of a commit",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            },
            "revision": {
                "type": "string",
                "description": "The revision (commit hash, branch name, tag) to show"
            }
        },
        "required": ["repo_path", "revision"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath string `json:"repo_path"`
			Revision string `json:"revision"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "git", "-C", input.RepoPath, "show", input.Revision)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git show error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitInitTool = mcp.Tool{
	Name:        "git_init",
	Description: "Initializes a Git repository",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "repo_path": {
                "type": "string",
                "description": "Path to directory to initialize git repo"
            }
        },
        "required": ["repo_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoPath string `json:"repo_path"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "git", "init", input.RepoPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git init error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitCloneTool = mcp.Tool{
	Name:        "git_clone",
	Description: "Clones a Git repository to a target directory",
	InputSchema: json.RawMessage(`{
		"type": "object",
		"properties": {
			"repo_url": {
				"type": "string",
				"description": "URL of the Git repository to clone"
			},
			"target_dir": {
				"type": "string",
				"description": "Path to the target directory where the repository will be cloned"
			}
		},
		"required": ["repo_url", "target_dir"]
	}`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			RepoURL   string `json:"repo_url"`
			TargetDir string `json:"target_dir"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		cmd := exec.CommandContext(ctx, "git", "clone", input.RepoURL, input.TargetDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git clone error: %w", err)
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{Type: "text", Text: string(output)}},
		}, nil
	},
}

var gitReadLocalFileTool = mcp.Tool{
	Name:        "local_git_read_file",
	Description: "Reads a specific file from the local Git repository and returns its content with line numbers",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "file_path": {
                "type": "string",
                "description": "Path to the file within the local repository"
            },
            "ref": {
                "type": "string",
                "description": "Git reference (branch, commit, or tag). If empty, reads from working directory",
                "default": ""
            }
        },
        "required": ["file_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			FilePath string `json:"file_path"`
			Ref      string `json:"ref"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		var content []byte
		if input.Ref != "" {
			// Read file from specific git reference
			cmd := exec.CommandContext(ctx, "git", "show", fmt.Sprintf("%s:%s", input.Ref, input.FilePath))
			content, err = cmd.Output()
			if err != nil {
				return mcp.CallToolResult{}, fmt.Errorf("failed to read file from git: %w", err)
			}
		} else {
			// Read file from working directory
			content, err = os.ReadFile(input.FilePath)
			if err != nil {
				return mcp.CallToolResult{}, fmt.Errorf("failed to read file: %w", err)
			}
		}

		// Split content into lines and add line numbers
		lines := strings.Split(string(content), "\n")
		var numberedLines []string
		for i, line := range lines {
			numberedLines = append(numberedLines, fmt.Sprintf("%d. %s", i+1, line))
		}

		// Format the output
		formattedContent := fmt.Sprintf("FILE: %s\nCONTENT:\n<<<\n%s\n>>>",
			input.FilePath,
			strings.Join(numberedLines, "\n"))

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{
				Type: "text",
				Text: formattedContent,
			}},
		}, nil
	},
}

var gitApplyPatchTool = mcp.Tool{
	Name:        "local_git_apply_patch",
	Description: "Applies a git patch to the local Git repository. Can handle both unified diff format and git patch format.",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "patch_content": {
                "type": "string",
                "description": "Content of the patch to apply"
            },
            "options": {
                "type": "object",
                "description": "Optional git apply parameters",
                "properties": {
                    "check_only": {
                        "type": "boolean",
                        "description": "Only check if patch can be applied, don't actually apply it",
                        "default": false
                    },
                    "reject": {
                        "type": "boolean",
                        "description": "Create .rej files for rejects",
                        "default": false
                    },
                    "reverse": {
                        "type": "boolean",
                        "description": "Apply patch in reverse",
                        "default": false
                    }
                }
            }
        },
        "required": ["patch_content"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			PatchContent string `json:"patch_content"`
			Options      struct {
				CheckOnly bool `json:"check_only"`
				Reject    bool `json:"reject"`
				Reverse   bool `json:"reverse"`
			} `json:"options"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("failed to unmarshal input: %w", err)
		}

		// Create a temporary file for the patch
		tmpFile, err := os.CreateTemp("", "git-patch-*.patch")
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("failed to create temporary patch file: %w", err)
		}
		//defer os.Remove(tmpFile.Name())

		// Write patch content to temporary file
		if _, err := tmpFile.WriteString(input.PatchContent); err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("failed to write patch content: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("failed to close temporary file: %w", err)
		}

		// Prepare git apply command with options
		args := []string{"apply"}

		if input.Options.CheckOnly {
			args = append(args, "--check")
		}
		if input.Options.Reject {
			args = append(args, "--reject")
		}
		if input.Options.Reverse {
			args = append(args, "-R")
		}

		args = append(args, tmpFile.Name())

		cmd := exec.CommandContext(ctx, "git", args...)
		output, err := cmd.CombinedOutput()

		log.Printf("Output tmp patch file: %s", tmpFile.Name())
		log.Printf("cmd: %+v", args)

		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("failed to apply patch: %w\nOutput: %s", err, string(output))
		}

		resultMsg := "Patch applied successfully"
		if input.Options.CheckOnly {
			resultMsg = "Patch can be applied successfully"
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{
				Type: "text",
				Text: fmt.Sprintf("%s\nCommand output:\n%s", resultMsg, string(output)),
			}},
		}, nil
	},
}

var gitAllInOneTool = mcp.Tool{
	Name:        "git_all_in_one",
	Description: "Performs any Git operation based on the provided command",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "command": {
                "type": "string",
                "description": "Git command to execute"
            },
            "repo_path": {
                "type": "string",
                "description": "Path to Git repository"
            },
            "args": {
                "type": "array",
                "items": {
                    "type": "string"
                },
                "description": "Arguments for the Git command"
            }
        },
        "required": ["command", "repo_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		span.SetAttributes(
			attribute.String("tool_name", params.Name),
			attribute.String("tool_argument", string(params.Arguments)),
		)
		defer span.End()

		var err error
		defer func() {
			if err != nil {
				span.RecordError(err)
			}
		}()

		var input struct {
			Command  string   `json:"command"`
			RepoPath string   `json:"repo_path"`
			Args     []string `json:"args"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("failed to unmarshal input: %w", err)
		}

		args := append([]string{"-C", input.RepoPath, input.Command}, input.Args...)
		cmd := exec.CommandContext(ctx, "git", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("git %s error: %w\nOutput: %s", input.Command, err, string(output))
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{{
				Type: "text",
				Text: string(output),
			}},
		}, nil
	},
}
