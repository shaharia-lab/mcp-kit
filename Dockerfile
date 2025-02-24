# Stage 1: Build frontend and backend
# Change from golang:1.23-alpine to use a Node.js 20 base image first
FROM node:20-alpine AS frontend-builder

# Set working directory
WORKDIR /app

# Copy package files first to leverage Docker cache
COPY mcp-frontend/package*.json ./mcp-frontend/

# Install frontend dependencies
RUN cd mcp-frontend && npm ci

# Copy the rest of the frontend files
COPY mcp-frontend ./mcp-frontend/

# Build frontend
RUN cd mcp-frontend && npm run build

# Now use golang image for backend
FROM golang:1.23-alpine AS backend-builder

# Install necessary build tools
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Create a non-root user
RUN adduser -D -g '' app

# Copy the entire project including the built frontend
COPY --from=frontend-builder /app/mcp-frontend/dist ./mcp-frontend/dist
COPY . .

# Build backend
RUN make build

# Stage 3: Create the final image
FROM alpine:3.19

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Import the user and group files from builder
COPY --from=backend-builder /etc/passwd /etc/passwd

# Copy the binary and static files from builder
COPY --from=backend-builder /app/build/mcp /usr/local/bin/mcp
COPY --from=backend-builder /app/cmd/static /usr/local/bin/static

# Use non-root user
USER app

# Set environment variables
ENV TZ=UTC \
    APP_USER=app \
    APP_PORT=8080

# Expose both API and frontend ports
EXPOSE 8080 8081

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/mcp"]

# Default command (can be overridden)
CMD []