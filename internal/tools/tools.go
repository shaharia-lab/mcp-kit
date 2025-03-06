package tools

import (
	"github.com/shaharia-lab/goai/mcp"
)

var MCPToolsRegistry = []mcp.Tool{
	weatherTool,
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
	curlAllInOneTool,
	findContentInFilesTool,
	findLargeFilesTool,
	findDuplicateFilesTool,
	findRecentlyModifiedTool,
	searchCodePatternTool,
}
