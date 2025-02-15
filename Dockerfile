# syntax=docker/dockerfile:1

# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Create a non-root user
RUN adduser -D -g '' app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/mcp

# Stage 2: Create the final image
FROM alpine:3.19

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Import the user and group files from builder
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary from builder
COPY --from=builder /go/bin/mcp /usr/local/bin/mcp

# Use non-root user
USER app

# Set environment variables
ENV TZ=UTC \
    APP_USER=app \
    APP_PORT=8080

# Expose the application port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/mcp"]

# Default command (can be overridden)
CMD []