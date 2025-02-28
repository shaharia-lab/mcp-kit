package prompt

import "github.com/shaharia-lab/goai/mcp"

var MCPPromptsRegistry = []mcp.Prompt{
	PromptLLMWithToolsUsage,
	PromptLLMWithToolsUsageV2,
	PromptLLMWithToolsUsageV3,
	PromptLLMGeneralMarkdown,
}
