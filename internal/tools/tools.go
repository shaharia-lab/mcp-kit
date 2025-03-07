package tools

import (
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/mcp-tools"
)

var MCPToolsRegistry = []mcp.Tool{
	mcptools.GetWeather,
	// Git tools
	gitAllInOneTool,
	// GitHub tools
	githubCreateRepository,
	githubCreateIssue,
	githubGetFileContents,
	githubCreateOrUpdateFile,
	githubPushFiles,
	githubSearchRepositories,
	githubCreatePullRequest,
	githubForkRepository,
	githubCreateBranch,
	githubListIssues,
	githubUpdateIssue,
	githubAddIssueComment,
	githubSearchCode,
	githubSearchIssues,
	githubSearchUsers,
	githubListCommits,
	githubGetIssue,
	githubGetPullRequest,
	githubListPullRequests,
	githubCreatePullRequestReview,
	githubMergePullRequest,
	githubGetPullRequestFiles,
	githubGetPullRequestStatus,
	githubUpdatePullRequestBranch,
	githubGetPullRequestComments,
	githubGetPullRequestReviews,
	// Filesystem tools
	fileSystemGetFileInfo,
	fileSystemListDirectory,
	fileSystemReadFile,
	fileSystemWriteFile,
	// PostgreSQL tools
	postgresExecuteQuery,
	postgresTableSchema,
	postgresExecuteQueryWithExplain,
	// Docker tools
	dockerAllInOneTool,
	findContentInFilesTool,
	findLargeFilesTool,
	findDuplicateFilesTool,
	findRecentlyModifiedTool,
	searchCodePatternTool,
}
