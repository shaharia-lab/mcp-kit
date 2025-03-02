package prompt

import "github.com/shaharia-lab/goai/mcp"

var MCPPromptsRegistry = []mcp.Prompt{
	PromptLLMWithToolsUsage,
	PromptLLMWithToolsUsageV2,
	PromptLLMWithToolsUsageV2UseChatHistory,
	PromptLLMWithToolsUsageV3,
	PromptLLMGeneralMarkdown,
}
