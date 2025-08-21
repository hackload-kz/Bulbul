# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates for go mod download
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the API server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o api-server ./cmd/api

# Build the consumers
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o consumers ./cmd/consumers

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/api-server /app/api-server
COPY --from=builder /app/consumers /app/consumers

# Change ownership of binaries
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (default API port)
EXPOSE 8081

# Default command runs the API server
CMD ["/app/api-server"]