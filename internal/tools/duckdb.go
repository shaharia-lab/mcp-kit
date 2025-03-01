package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/shaharia-lab/goai/mcp"
	"github.com/shaharia-lab/goai/observability"
)

// Tool for listing all tables
var duckdbListTables = mcp.Tool{
	Name:        "duckdb_list_tables",
	Description: "Lists all tables in the DuckDB database.",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "database_path": {
                "type": "string",
                "description": "Path to the DuckDB database file"
            }
        },
        "required": ["database_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		defer span.End()

		var input struct {
			DatabasePath string `json:"database_path"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		db, err := sql.Open("duckdb", input.DatabasePath)
		if err != nil {
			return mcp.CallToolResult{}, err
		}
		defer db.Close()

		rows, err := db.QueryContext(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = 'main'")
		if err != nil {
			return mcp.CallToolResult{}, err
		}
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return mcp.CallToolResult{}, err
			}
			tables = append(tables, tableName)
		}

		tableList, _ := json.Marshal(tables)
		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: string(tableList),
				},
			},
		}, nil
	},
}

// Tool for getting schema information
var duckdbGetSchema = mcp.Tool{
	Name:        "duckdb_get_schema",
	Description: "Get schema information for one or all tables in DuckDB.",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "database_path": {
                "type": "string",
                "description": "Path to the DuckDB database file"
            },
            "table_name": {
                "type": "string",
                "description": "Optional: Specific table name. If not provided, returns schema for all tables"
            }
        },
        "required": ["database_path"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		defer span.End()

		var input struct {
			DatabasePath string `json:"database_path"`
			TableName    string `json:"table_name,omitempty"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		db, err := sql.Open("duckdb", input.DatabasePath)
		if err != nil {
			return mcp.CallToolResult{}, err
		}
		defer db.Close()

		query := `
            SELECT 
                table_name,
                column_name,
                data_type,
                is_nullable
            FROM information_schema.columns 
            WHERE table_schema = 'main'
        `
		if input.TableName != "" {
			query += fmt.Sprintf(" AND table_name = '%s'", input.TableName)
		}

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return mcp.CallToolResult{}, err
		}
		defer rows.Close()

		schemas := make(map[string][]map[string]string)
		for rows.Next() {
			var tableName, columnName, dataType, isNullable string
			if err := rows.Scan(&tableName, &columnName, &dataType, &isNullable); err != nil {
				return mcp.CallToolResult{}, err
			}

			columnInfo := map[string]string{
				"column_name": columnName,
				"data_type":   dataType,
				"nullable":    isNullable,
			}
			schemas[tableName] = append(schemas[tableName], columnInfo)
		}

		schemaJSON, _ := json.MarshalIndent(schemas, "", "  ")
		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: string(schemaJSON),
				},
			},
		}, nil
	},
}

// Tool for executing queries
var duckdbExecuteQuery = mcp.Tool{
	Name:        "duckdb_execute_query",
	Description: "Execute a SQL query in DuckDB and return the results.",
	InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "database_path": {
                "type": "string",
                "description": "Path to the DuckDB database file"
            },
            "query": {
                "type": "string",
                "description": "SQL query to execute"
            }
        },
        "required": ["database_path", "query"]
    }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		ctx, span := observability.StartSpan(ctx, fmt.Sprintf("%s.Handler", params.Name))
		defer span.End()

		var input struct {
			DatabasePath string `json:"database_path"`
			Query        string `json:"query"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		db, err := sql.Open("duckdb", input.DatabasePath)
		if err != nil {
			return mcp.CallToolResult{}, err
		}
		defer db.Close()

		rows, err := db.QueryContext(ctx, input.Query)
		if err != nil {
			return mcp.CallToolResult{}, err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return mcp.CallToolResult{}, err
		}

		var results []map[string]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return mcp.CallToolResult{}, err
			}

			row := make(map[string]interface{})
			for i, col := range columns {
				row[col] = values[i]
			}
			results = append(results, row)
		}

		resultJSON, _ := json.MarshalIndent(results, "", "  ")
		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: string(resultJSON),
				},
			},
		}, nil
	},
}
