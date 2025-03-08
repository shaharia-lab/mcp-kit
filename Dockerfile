FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git make

WORKDIR /app
RUN adduser -D -g '' app

COPY . .
RUN make build-in-docker

# Final stage
FROM alpine:3.19

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Import the user and group files from builder
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary from builder
COPY --from=builder /app/build/mcp /usr/local/bin/mcp

# Use non-root user
USER app

# Set environment variables
ENV TZ=UTC \
    APP_USER=app

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/mcp"]

# Default command (can be overridden)
CMD []