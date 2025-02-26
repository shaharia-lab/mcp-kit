# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app

COPY mcp-frontend/package*.json ./mcp-frontend/
RUN cd mcp-frontend && npm ci
COPY mcp-frontend ./mcp-frontend/
RUN cd mcp-frontend && npm run build

FROM golang:1.23-alpine AS backend-builder
RUN apk add --no-cache git make

WORKDIR /app
RUN adduser -D -g '' app

COPY . .
COPY --from=frontend-builder /app/cmd/static ./cmd/static

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
    APP_USER=app

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/mcp"]

# Default command (can be overridden)
CMD []