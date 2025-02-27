package prompt

import "github.com/shaharia-lab/goai/mcp"

var PromptLLMWithToolsUsage = mcp.Prompt{
	Name:        "llm_with_tools",
	Description: "Ask LLM question and provide access to several tools and return response in Markdown",
	Arguments: []mcp.PromptArgument{
		{
			Name:        "question",
			Description: "Question asked by the user",
			Required:    true,
		},
	},
	Messages: []mcp.PromptMessage{
		{
			Role: "user",
			Content: mcp.PromptContent{
				Type: "text",
				Text: `You are a helpful AI assistant with access to various tools and functions.

# Response Format
- Provide concise answers upfront with minimal preamble
- Use proper Markdown formatting (code blocks with language specs, links when needed)
- Place citations/references at the bottom using numbered footnotes

# Guidelines
- Execute tools in proper sequence, waiting for results before dependent calls
- Link assertions in your answer to numbered sources using [^1] format
- Keep responses brief and direct

# Clarification Protocol
- If request is ambiguous, ask specific questions before proceeding
- Never make assumptions about unclear requirements

# Example Format
[Direct answer with citation references][^1]

[^1]: [Source description/tool used]

---
Question: {{question}}
`,
			},
		},
	},
}

var PromptLLMGeneralMarkdown = mcp.Prompt{
	Name:        "llm_general",
	Description: "Ask LLM question and return response in Markdown",
	Arguments: []mcp.PromptArgument{
		{
			Name:        "question",
			Description: "Question asked by the user",
			Required:    true,
		},
	},
	Messages: []mcp.PromptMessage{
		{
			Role: "user",
			Content: mcp.PromptContent{
				Type: "text",
				Text: `You are a helpful AI assistant.

# Response Format
- Present direct answers with minimal introduction
- Use proper Markdown formatting (code blocks with language, lists, headers)
- Organize content with clear section breaks

# Guidelines
- Keep responses focused and concise
- Support answers with relevant examples
- Maintain consistent formatting style

# Clarification Protocol
- Ask specific questions for unclear requests
- One clarification at a time
- No assumptions about ambiguous requirements

# Example Format
[Clear, direct response with proper formatting]

---
Question: {{question}}`,
			},
		},
	},
}
