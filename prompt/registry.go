package prompt

import "github.com/shaharia-lab/goai/mcp"

var MCPPromptsRegistry = []mcp.Prompt{
	PromptLLMWithToolsUsage,
	PromptLLMGeneralMarkdown,
}
