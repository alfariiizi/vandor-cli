# Build stage
FROM golang:1.24-alpine AS builder

# Install git for go mod
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X 'main.version=docker' -X 'main.commit=$(git rev-parse --short HEAD)' -X 'main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
    -o vandor main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS and git for source upgrades
RUN apk --no-cache add ca-certificates git

# Create non-root user
RUN adduser -D -s /bin/sh vandor

# Set working directory
WORKDIR /home/vandor

# Copy binary from builder stage
COPY --from=builder /app/vandor /usr/local/bin/vandor

# Make sure binary is executable
RUN chmod +x /usr/local/bin/vandor

# Switch to non-root user
USER vandor

# Set entrypoint
ENTRYPOINT ["vandor"]

# Default command
CMD ["--help"]