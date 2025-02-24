package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/observability"
	"net/http"
)

// ToolInfo represents a simplified structure for tools
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Handler to list all tools
func ListToolsHandler(toolsProvider *goai.ToolsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start a new observability span
		ctx, span := observability.StartSpan(r.Context(), "handle_list_tools")
		defer span.End()

		// Get the tools from the provider
		tools, err := toolsProvider.ListTools(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get tools: %v", err), http.StatusInternalServerError)
			span.RecordError(err)
			return
		}

		// Convert tools to a simplified response format
		toolInfos := make([]ToolInfo, len(tools))
		for i, tool := range tools {
			toolInfos[i] = ToolInfo{
				Name:        tool.Name,
				Description: tool.Description,
			}
		}

		// Return the response as JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(toolInfos); err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
			span.RecordError(err)
		}
	}
}
