# syntax=docker/dockerfile:1

# Stage 1: Build frontend and backend
FROM golang:1.23-alpine AS builder

# Install necessary build tools and Node.js
RUN apk add --no-cache git make nodejs npm

# Set working directory
WORKDIR /app

# Create a non-root user
RUN adduser -D -g '' app

# Copy the entire project
COPY . .

# Install frontend dependencies and build both frontend and backend
RUN cd mcp-frontend && npm ci && cd .. && make frontend-build && make build

# Stage 2: Create the final image
FROM alpine:3.19

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Import the user and group files from builder
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary and static files from builder
COPY --from=builder /app/build/mcp /usr/local/bin/mcp
COPY --from=builder /app/cmd/static /usr/local/bin/static

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