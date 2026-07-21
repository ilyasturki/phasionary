# The repo-root VERSION file is the single source of truth for the version,
# shared with the Go build (injected via ldflags below) and flake.nix, which
# stamps the same `v`-prefixed value.
version := trim(`cat VERSION`)
commit := `git rev-parse --short HEAD`
build_date := `date -u +%Y-%m-%dT%H:%M:%SZ`
ldflags := "-X phasionary/internal/version.Version=v" + version + \
    " -X phasionary/internal/version.Commit=" + commit + \
    " -X phasionary/internal/version.BuildDate=" + build_date

# Build the CLI binary.
build:
    go build -ldflags "{{ldflags}}" -o phasionary ./cmd/phasionary

# Run from source, passing args through (e.g. `just run version`).
run *ARGS:
    go run -ldflags "{{ldflags}}" ./cmd/phasionary {{ARGS}}

# Run from source against the local ./data dir.
run-app *ARGS:
    go run -ldflags "{{ldflags}}" ./cmd/phasionary --data ./data {{ARGS}}

# Run the compiled binary, building it first if needed.
exec *ARGS: build
    ./phasionary {{ARGS}}

# Serve the web UI with token auth (random token unless one is given).
serve token=`openssl rand -hex 16` host="127.0.0.1" port="7777": build
    @echo "Web UI: http://{{host}}:{{port}}/?token={{token}}"
    ./phasionary serve --host "{{host}}" --port "{{port}}" --token "{{token}}"

# Build the CLI binary using Nix.
build-nix:
    nix build

# Run the CLI binary using Nix, passing args through (e.g. `just run-nix version`).
run-nix *ARGS:
    nix run . -- {{ARGS}}

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
