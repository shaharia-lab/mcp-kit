# syntax=docker/dockerfile:1

# Stage 1: Build the frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy frontend directory with package files
COPY ./mcp-frontend/package.json ./mcp-frontend/package-lock.json ./
COPY ./mcp-frontend ./mcp-frontend

# Install dependencies and build frontend
WORKDIR /app/mcp-frontend
RUN npm ci
RUN npm run build

# Stage 2: Build the Go application
FROM golang:1.23-alpine AS backend-builder

# Install necessary build tools and make
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Create a non-root user
RUN adduser -D -g '' app

# Copy go mod files
COPY go.mod go.sum ./
COPY Makefile ./

# Download dependencies
RUN go mod download

# Copy source code and the frontend build
COPY . .
COPY --from=frontend-builder /app/mcp-frontend/dist ./cmd/static

# Build the application using Make
RUN make build

# Stage 3: Create the final image
FROM alpine:3.19

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Import the user and group files from builder
COPY --from=backend-builder /etc/passwd /etc/passwd

# Copy the binary from builder (updated path)
COPY --from=backend-builder /app/build/mcp /usr/local/bin/mcp
COPY --from=backend-builder /app/cmd/static /usr/local/bin/static

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