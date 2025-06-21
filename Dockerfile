# Build stage
FROM golang:1.24.4-alpine AS builder

# Install git and ca-certificates (needed for go modules)
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Verify dependencies
RUN go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo -o score-checker ./cmd/score-checker

# Final stage - minimal image
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary
COPY --from=builder /build/score-checker /usr/local/bin/score-checker

# Use unprivileged user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/score-checker", "--help"]

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/score-checker"]

# Default command
CMD ["--help"]

# Labels for metadata
LABEL org.opencontainers.image.title="Score Checker"
LABEL org.opencontainers.image.description="A microservice that monitors Sonarr episodes and Radarr movies for low custom format scores"
LABEL org.opencontainers.image.vendor="Score Checker Project"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/yourusername/score-checker"
LABEL org.opencontainers.image.documentation="https://github.com/yourusername/score-checker/blob/main/README.md"