# MCP Kit (Model Context Protocol Kit)

## Installation

### Using Docker

```bash
# Build the image
docker build -t mcp:latest .

# Run server
docker run -d \
  --name mcp-server \
  -p 8080:8080 \
  -e MCP_SERVER_PORT=8080 \
  mcp:latest server

# Run client
docker run -d \
  --name mcp-client \
  --add-host=host.docker.internal:host-gateway \
  -e MCP_SERVER_URL=http://host.docker.internal:8080/events \
  mcp:latest client
```