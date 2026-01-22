# Build the CLI binary.
build:
    go build -o phasionary ./cmd/phasionary

# Run the CLI binary with the test data
run:
    ./phasionary --data ./data

# Run the CLI binary with the default data
run-default:
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
