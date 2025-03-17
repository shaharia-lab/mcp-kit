package tools

import (
	"fmt"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/goai/observability"
	mcptools "github.com/shaharia-lab/mcp-tools"
	"google.golang.org/api/gmail/v1"
)

// Registry is a collection of tools that can be used by the MCP.
type Registry struct {
	tools        map[string]mcp.Tool
	config       *ToolsConfig
	logger       observability.Logger
	gmailService *gmail.Service
}

// NewRegistry creates a new Registry instance.
func NewRegistry(cfg *ToolsConfig, logger observability.Logger, gmailService *gmail.Service) *Registry {
	return &Registry{
		tools:        make(map[string]mcp.Tool),
		config:       cfg,
		logger:       logger,
		gmailService: gmailService,
	}
}

// Init initializes the registry with a tool.
func (r *Registry) Init() error {
	if r.config.Weather.IsEnabled() {
		r.AddTool(mcptools.GetWeather.Name, mcptools.GetWeather)
	}

	if r.config.Docker.IsEnabled() {
		docker := mcptools.NewDocker(r.logger)
		r.AddTool(mcptools.DockerToolName, docker.DockerAllInOneTool())
	}

	if r.config.Git.IsEnabled() {
		git := mcptools.NewGit(r.logger, mcptools.GitConfig{
			DefaultRepoPath: r.config.Git.DefaultRepoPath,
			BlockedCommands: r.config.Git.BlockedCommands,
		})

		r.AddTool(mcptools.GitToolName, git.GitAllInOneTool())
	}

	if r.config.Filesystem.IsEnabled() {
		fileSystem := mcptools.NewFileSystem(r.logger, mcptools.FileSystemConfig{
			AllowedDirectory: r.config.Filesystem.AllowedDirectory,
			BlockedPatterns:  r.config.Filesystem.BlockedPattern,
		})
		r.AddTool(mcptools.FileSystemToolName, fileSystem.FileSystemAllInOneTool())
	}

	if r.config.Curl.IsEnabled() {
		curl := mcptools.NewCurl(r.logger, mcptools.CurlConfig{
			BlockedMethods: r.config.Curl.BlockedMethods,
		})
		r.AddTool(mcptools.CurlToolName, curl.CurlAllInOneTool())
	}

	if r.config.Postgres.IsEnabled() {
		postgres := mcptools.NewPostgreSQL(r.logger, mcptools.PostgreSQLConfig{
			DefaultDatabase: r.config.Postgres.Databases[0].Name,
			BlockedCommands: r.config.Postgres.Databases[0].BlockedCommands,
		})
		r.AddTool(mcptools.PostgreSQLToolName, postgres.PostgreSQLAllInOneTool())
	}

	if r.config.Bash.IsEnabled() {
		bash := mcptools.NewBash(r.logger)
		r.AddTool(mcptools.BashToolName, bash.BashAllInOneTool())
	}

	if r.config.Sed.IsEnabled() {
		sed := mcptools.NewSed(r.logger)
		r.AddTool(mcptools.SedToolName, sed.SedAllInOneTool())
	}

	if r.config.Grep.IsEnabled() {
		grep := mcptools.NewGrep(r.logger)
		r.AddTool(mcptools.GrepToolName, grep.GrepAllInOneTool())
	}

	if r.config.Cat.IsEnabled() {
		cat := mcptools.NewCat(r.logger)
		r.AddTool(mcptools.CatToolName, cat.CatAllInOneTool())
	}

	if r.config.GithubRepository.IsEnabled() {
		ghConfig := mcptools.NewGitHubTool(r.logger, mcptools.GitHubConfig{
			Token: r.config.GithubRepository.Token,
		})
		r.AddTool(mcptools.GitHubRepositoryToolName, ghConfig.GetRepositoryTool())
	}

	if r.config.GithubIssues.IsEnabled() {
		ghConfig := mcptools.NewGitHubTool(r.logger, mcptools.GitHubConfig{
			Token: r.config.GithubIssues.Token,
		})
		r.AddTool(mcptools.GitHubIssuesToolName, ghConfig.GetIssuesTool())
	}

	if r.config.GithubPulls.IsEnabled() {
		ghConfig := mcptools.NewGitHubTool(r.logger, mcptools.GitHubConfig{
			Token: r.config.GithubPulls.Token,
		})
		r.AddTool(mcptools.GitHubPullRequestsToolName, ghConfig.GetPullRequestsTool())
	}

	if r.config.GithubSearch.IsEnabled() {
		ghConfig := mcptools.NewGitHubTool(r.logger, mcptools.GitHubConfig{
			Token: r.config.GithubSearch.Token,
		})
		r.AddTool(mcptools.GitHubSearchToolName, ghConfig.GetSearchTool())
	}

	if r.config.Gmail.IsEnabled() {
		if r.gmailService == nil {
			r.logger.Warn("Gmail service is not set. Gmail tools might not work as expected.")
			return fmt.Errorf("gmail service is not set")
		}

		gmailTool := mcptools.NewGmail(r.logger, r.gmailService, mcptools.GmailConfig{
			UserID:         r.config.Gmail.UserID,
			MaxResults:     r.config.Gmail.MaxResults,
			SinceLastNDays: r.config.Gmail.SinceLastNDays,
		})
		r.AddTool(mcptools.GmailToolName, gmailTool.GmailAllInOneTool())
	}

	r.logger.WithFields(map[string]interface{}{"total_tools": len(r.tools)}).Info("All tools have been initialized")
	return nil
}

// GetTool retrieves a tool from the registry.
func (r *Registry) GetTool(toolID string) mcp.Tool {
	return r.tools[toolID]
}

// AddTool adds a tool to the registry.
func (r *Registry) AddTool(toolID string, tool mcp.Tool) {
	r.tools[toolID] = tool
}

// GetToolLists retrieves all tools from the registry.
func (r *Registry) GetToolLists() []mcp.Tool {
	var tools []mcp.Tool
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}
