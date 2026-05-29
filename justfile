# Build the CLI binary.
build:
    go build -ldflags "-X phasionary/internal/version.Version=$(git describe --tags --always) -X phasionary/internal/version.Commit=$(git rev-parse --short HEAD) -X phasionary/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o phasionary ./cmd/phasionary

# Run the CLI binary
run *ARGS:
    ./phasionary {{ARGS}}

# Run the CLI binary with the app data
run-app *ARGS:
    ./phasionary --data ./data {{ARGS}}

# Build the CLI binary using Nix.
build-nix:
    nix build

# Run the CLI binary using Nix.
run-nix:
    nix run

# Run all tests.
test:
    go test ./...
    nix build

# Format Go files.
fmt:
    gofmt -w cmd/phasionary internal

# Clean up go.mod/go.sum.
tidy:
    go mod tidy

# Bump version (X.Y.Z or major|minor|patch); updates VERSION, commits, tags.
bump version:
    ./scripts/bump.sh {{version}}
