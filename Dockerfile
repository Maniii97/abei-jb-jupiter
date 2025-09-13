# Build stage
FROM golang:1.23.2-alpine AS builder

# Install git and ca-certificates (for fetching dependencies and making HTTPS requests)
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Create appuser for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main \
    ./cmd/api

# Final stage
FROM scratch

# Import ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the user and group files from the builder
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary from builder
COPY --from=builder /build/main /app/

# Use an unprivileged user
USER appuser

# Expose port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/main"]
