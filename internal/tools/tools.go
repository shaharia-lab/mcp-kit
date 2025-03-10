package tools

import (
	"github.com/shaharia-lab/goai/mcp"
	mcptools "github.com/shaharia-lab/mcp-tools"
)

var MCPToolsRegistry = []mcp.Tool{
	mcptools.GetWeather,
}
