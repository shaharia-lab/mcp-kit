package tools

import (
	"github.com/shaharia-lab/goai/mcp"
)

var MCPToolsRegistry = []mcp.Tool{
	weatherTool,
	// Git tools
	gitStatusTool,
	/*gitDiffUnstagedTool,
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
	// Filesystem tools
	fileSystemGetFileInfo,
	fileSystemListDirectory,
	fileSystemReadFile,
	fileSystemWriteFile,
	// PostgreSQL tools
	postgresExecuteQuery,
	postgresTableSchema,
	postgresExecuteQueryWithExplain,*/
}
