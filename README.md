# MCP Kit (Model Context Protocol Kit)

## Installation

### Using Docker

```bash
# Build the image
docker build -t mcp:latest .

# Run server
docker run -p 8080:8080 mcp:latest server

# Run client
docker run mcp:latest client --server http://host.docker.internal:8080
docker run -d --name mcp-server -p 8080:8080 mcp:latest server
```