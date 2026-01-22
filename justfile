# Build the CLI binary.
build:
    go build -o phasionary ./cmd/phasionary

# Run the CLI binary.
run:
    ./phasionary --data ./data

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
