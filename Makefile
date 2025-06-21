# Score Checker Makefile

.PHONY: build test test-verbose test-coverage benchmark clean lint vet fmt docker-build docker-run docker-push

# Build the application
build:
	go build -o score-checker ./cmd/score-checker

# Run all tests
test:
	go test ./internal/...

# Run tests with verbose output
test-verbose:
	go test -v ./internal/...

# Run tests with coverage
test-coverage:
	go test -cover ./internal/...

# Generate detailed coverage report
test-coverage-html:
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated in coverage.html"

# Run benchmarks
benchmark:
	go test -bench=. -benchmem ./internal/...

# Run linter (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Run go vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -f score-checker coverage.out coverage.html
	docker image prune -f

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run all quality checks
quality: fmt vet test-coverage

# Development setup
dev-setup:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker build
docker-build:
	docker build -t score-checker:latest .

# Docker run (interactive)
docker-run: docker-build
	docker run --rm -it -v $(PWD)/config.yaml:/app/config.yaml score-checker:latest

# Docker push to GHCR (requires authentication)
docker-push: docker-build
	docker tag score-checker:latest ghcr.io/yourusername/score-checker:latest
	docker push ghcr.io/yourusername/score-checker:latest

# Build multi-platform release binaries
build-release:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dist/score-checker-linux-amd64 ./cmd/score-checker
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o dist/score-checker-linux-arm64 ./cmd/score-checker
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o dist/score-checker-darwin-amd64 ./cmd/score-checker
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o dist/score-checker-darwin-arm64 ./cmd/score-checker
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o dist/score-checker-windows-amd64.exe ./cmd/score-checker

# Show help
help:
	@echo "Available targets:"
	@echo "  build              Build the application"
	@echo "  test               Run all tests"
	@echo "  test-verbose       Run tests with verbose output"
	@echo "  test-coverage      Run tests with coverage"
	@echo "  test-coverage-html Generate HTML coverage report"
	@echo "  benchmark          Run benchmarks"
	@echo "  lint              Run golangci-lint"
	@echo "  vet               Run go vet"
	@echo "  fmt               Format code"
	@echo "  clean             Clean build artifacts"
	@echo "  deps              Download and tidy dependencies"
	@echo "  quality           Run fmt, vet, and test-coverage"
	@echo "  dev-setup         Install development dependencies"
	@echo "  docker-build      Build Docker image"
	@echo "  docker-run        Build and run Docker container"
	@echo "  docker-push       Push Docker image to GHCR"
	@echo "  build-release     Build multi-platform release binaries"
	@echo "  help              Show this help message"