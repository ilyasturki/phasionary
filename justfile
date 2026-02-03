# Build the CLI binary.
build:
    go build -o phasionary ./cmd/phasionary

# Run the CLI binary
run:
    ./phasionary

# Run the CLI binary with the app data
run-app:
    ./phasionary --data ./data

# Run all tests.
test:
    go test ./...

# Format Go files.
fmt:
    gofmt -w cmd/phasionary internal

# Clean up go.mod/go.sum.
tidy:
    go mod tidy
