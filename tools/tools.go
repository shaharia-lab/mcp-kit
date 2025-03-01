package tools

import (
	"github.com/shaharia-lab/goai/mcp"
)

var MCPToolsRegistry = []mcp.Tool{
	weatherTool,
	// Git tools
	gitStatusTool,
	gitDiffUnstagedTool,
	gitDiffStagedTool,
	gitDiffTool,
	gitCommitTool,
	gitAddTool,
	gitResetTool,
	gitLogTool,
	gitCreateBranchTool,
	gitCheckoutTool,
	gitShowTool,
	gitInitTool,
	gitCloneTool,
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
}
