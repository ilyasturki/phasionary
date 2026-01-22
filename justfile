set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

# Build the CLI binary.
build:
	go build -o phasionary ./cmd/phasionary

# Run all tests.
test:
	go test ./...

# Run tests for domain package only.
test-domain:
	go test -v ./internal/domain/...

# Format Go files.
fmt:
	gofmt -w cmd/phasionary internal

# Clean up go.mod/go.sum.
tidy:
	go mod tidy
