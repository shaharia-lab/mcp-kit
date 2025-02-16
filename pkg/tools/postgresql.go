package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/shaharia-lab/goai/mcp"
	"log"
	"strings"
)

// getDBConnection handles database connection configuration internally
func getDBConnection(dbName string) (*sql.DB, error) {
	dbConnectionStr := map[string]string{
		"app": fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			"localhost",
			"5432",
			"app",
			"pass",
			dbName,
		),
	}

	connStr, ok := dbConnectionStr[dbName]
	if !ok {
		return nil, fmt.Errorf("database not found: %s", dbName)
	}

	return sql.Open("postgres", connStr)
}

var postgresTableSchema = mcp.Tool{
	Name:        "postgresql_table_schema",
	Description: "Get the schema definition of a PostgreSQL table",
	InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "database_name": {
                    "type": "string",
                    "description": "Name of the database"
                },
                "table_name": {
                    "type": "string",
                    "description": "Name of the table to get schema for"
                }
            },
            "required": ["database_name", "table_name"]
        }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			DatabaseName string `json:"database_name"`
			TableName    string `json:"table_name"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		db, err := getDBConnection(input.DatabaseName)
		if err != nil {
			return mcp.CallToolResult{}, fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()

		query := `
                SELECT column_name, data_type, character_maximum_length, 
                       is_nullable, column_default
                FROM information_schema.columns 
                WHERE table_name = $1
                ORDER BY ordinal_position;
            `

		log.Printf("Query: %s", query)

		rows, err := db.QueryContext(ctx, query, input.TableName)
		if err != nil {
			return mcp.CallToolResult{}, err
		}
		defer rows.Close()

		var schema strings.Builder
		schema.WriteString(fmt.Sprintf("Table: %s\n\n", input.TableName))
		schema.WriteString("Column Name | Data Type | Length | Nullable | Default\n")
		schema.WriteString("------------|-----------|---------|----------|----------\n")

		for rows.Next() {
			var (
				columnName, dataType, isNullable string
				maxLength                        sql.NullInt64
				defaultValue                     sql.NullString
			)
			if err := rows.Scan(&columnName, &dataType, &maxLength, &isNullable, &defaultValue); err != nil {
				return mcp.CallToolResult{}, err
			}

			schema.WriteString(fmt.Sprintf("%s | %s | %v | %s | %s\n",
				columnName,
				dataType,
				maxLength.Int64,
				isNullable,
				defaultValue.String))
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: schema.String(),
				},
			},
		}, nil
	},
}

var postgresExecuteQuery = mcp.Tool{
	Name:        "postgresql_execute_query",
	Description: "Execute a PostgreSQL query and return the results",
	InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "database_name": {
                    "type": "string",
                    "description": "Name of the database"
                },
                "query": {
                    "type": "string",
                    "description": "SQL query to execute"
                }
            },
            "required": ["database_name", "query"]
        }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			DatabaseName string `json:"database_name"`
			Query        string `json:"query"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		db, err := getDBConnection(input.DatabaseName)
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

		var result strings.Builder
		result.WriteString(strings.Join(columns, " | ") + "\n")
		result.WriteString(strings.Repeat("-", len(strings.Join(columns, " | "))) + "\n")

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		for rows.Next() {
			if err := rows.Scan(valuePtrs...); err != nil {
				return mcp.CallToolResult{}, err
			}

			var rowValues []string
			for _, val := range values {
				rowValues = append(rowValues, fmt.Sprintf("%v", val))
			}
			result.WriteString(strings.Join(rowValues, " | ") + "\n")
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: result.String(),
				},
			},
		}, nil
	},
}

var postgresExecuteQueryWithExplain = mcp.Tool{
	Name:        "postgresql_execute_query_with_explain",
	Description: "Execute a PostgreSQL query with EXPLAIN ANALYZE",
	InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "database_name": {
                    "type": "string",
                    "description": "Name of the database"
                },
                "query": {
                    "type": "string",
                    "description": "SQL query to explain and execute"
                }
            },
            "required": ["database_name", "query"]
        }`),
	Handler: func(ctx context.Context, params mcp.CallToolParams) (mcp.CallToolResult, error) {
		var input struct {
			DatabaseName string `json:"database_name"`
			Query        string `json:"query"`
		}
		if err := json.Unmarshal(params.Arguments, &input); err != nil {
			return mcp.CallToolResult{}, err
		}

		db, err := getDBConnection(input.DatabaseName)
		if err != nil {
			return mcp.CallToolResult{}, err
		}
		defer db.Close()

		explainQuery := "EXPLAIN ANALYZE " + input.Query
		rows, err := db.QueryContext(ctx, explainQuery)
		if err != nil {
			return mcp.CallToolResult{}, err
		}
		defer rows.Close()

		var explain strings.Builder
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				return mcp.CallToolResult{}, err
			}
			explain.WriteString(line + "\n")
		}

		return mcp.CallToolResult{
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: explain.String(),
				},
			},
		}, nil
	},
}
