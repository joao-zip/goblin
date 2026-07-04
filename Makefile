.PHONY: test lint build clean

# Run all tests
test:
	go test ./... -v

# Run tests with coverage
cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Build the CLI binary
build:
	go build -o goblin ./cmd/goblin

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -f goblin coverage.out coverage.html

