# MCP Kit - Model Context Protocol Toolkit

## Overview

The MCP Kit provides a platform that facilitates interaction with Large Language Models (LLMs) using the Model Context Protocol (MCP).
It enables AI assistants to interact with external tools and services, extending their capabilities beyond their confined contexts.
This toolkit offers a standardized way for AI models to communicate with external systems.

## Key Features

*   **API Server**: Provides LLM endpoints to interact with AI models.
*   **MCP Client**: Configurable client to connect to the MCP server.
*   **Tools**: There are many tools already implemented to access external systems (i.e: filesystem, git, github, postgresql, etc..), and you can easily add more.
*   **Interactive Chat Interface**: Provides an interactive chat interface for users to interact with the AI model.
*   **Chat History**: Implements in-memory storage for maintaining chat history.
*   **LLM Providers**: Supports Anthropic, OpenAI, Amazon Bedrock, Cohere, Meta Llama, Mistral, and DeepSeek models. You can integrate more models easily.

<video src="https://github.com/user-attachments/assets/81804a29-e896-4f65-a929-05ac6a6aa92a" controls title="MCP Kit in action"></video>

## Architecture

The MCP Kit follows a client-server architecture:

1.  **Client**: Sends user requests to the MCP server.
2.  **Server**: Processes requests, coordinates with AI models, and manages tool execution.

### Request Flow

1.  The client sends a user query to the MCP server.
2.  The server passes the query to the LLM API.
3.  The LLM determines if tools are needed.
4.  If tools are needed:
    *   The LLM sends a request with tool parameters.
    *   The MCP server executes the tools.
    *   The tool output is returned to the LLM.
    *   This loop continues until no more tools are needed.
5.  The final response is sent back to the client.

## Tools

The kit includes a variety of tools:

*   **Git Tools**: `git_status`, `git_diff`, `git_commit`, `git_add`, `git_log`, etc..
*   **File System Tools**: `filesystem_list_directory`, `filesystem_read_file`, `filesystem_write_file`, `filesystem_get_file_info`.
*   **GitHub Tools**: `github_create_repository`, `github_create_issue`, `github_get_file_contents`, `github_search_repositories` etc.
*   **PostgreSQL Tools**: `postgresql_execute_query`, `postgresql_table_schema`, `postgresql_execute_query_with_explain`.

## Configuration

The configuration is loaded via environment variables.  Key configuration parameters include:

*   `API_SERVER_PORT`: Port for the API server (default: 8081).
*   `MCP_SERVER_URL`: URL for the MCP server (default: `http://localhost:8080/events`).
*   `MCP_SERVER_PORT`: Port for the MCP server (default: 8080).
*   `TOOLS_ENABLED`: List of enabled tools (default: `get_weather`).

## Getting Started

### Prerequisites

*   Go installed
*   [Include any other prerequisites]

### Installation

#### Load .env file

```bash
cp .env.example .env.local
export $(cat .env.local | grep -v '^#' | xargs)
```

#### Using Source Code

```bash
git clone git@github.com:shaharia-lab/mcp-kit.git
cd mcp-kit
make build
```

##### Running the MCP Server

```bash
./mcp server
```

##### Running the API Server

```bash
./mcp api
```

#### Using Docker

```bash
docker pull ghcr.io/shaharia-lab/mcp-kit:$VERSION

# Run MCP server
docker run -d \
  --name mcp-server \
  -p 8080:8080 \
  -e MCP_SERVER_PORT=8080 \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  ghcr.io/shaharia-lab/mcp-kit:$VERSION server

# Run API server
docker run -d \
  --name mcp-client \
  --add-host=host.docker.internal:host-gateway \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  -e MCP_SERVER_URL=http://host.docker.internal:8080/events \
  -p 8081:8081 \
  ghcr.io/shaharia-lab/mcp-kit:$VERSION api
```

Visit `http://localhost:8081` to access the UI interface to interact with the AI model.

### Interacting with the API

OpenAPI schema is available in `openapi.yaml`.