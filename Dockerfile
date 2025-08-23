# Build stage
FROM golang:1.23-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pr-compass ./cmd/pr-compass

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and git for GitHub CLI
RUN apk --no-cache add ca-certificates git curl

# Install GitHub CLI for Alpine
RUN apk add --no-cache github-cli

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/pr-compass .

# Create config directory
RUN mkdir -p /root/.config

# Expose port (if we add web features later)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./pr-compass --version || exit 1

# Run the binary
ENTRYPOINT ["./pr-compass"]
