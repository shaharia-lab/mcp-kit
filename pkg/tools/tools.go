package tools

import (
	"github.com/shaharia-lab/goai/mcp"
)

var MCPToolsRegistry = append([]mcp.Tool{
	weatherTool,
}, FilesystemTools...)
