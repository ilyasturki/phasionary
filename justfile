# The repo-root VERSION file is the single source of truth for the version,
# shared with the Go build (injected via ldflags below) and flake.nix, which
# stamps the same `v`-prefixed value.
version := trim(`cat VERSION`)
commit := `git rev-parse --short HEAD`
build_date := `date -u +%Y-%m-%dT%H:%M:%SZ`
ldflags := "-X phasionary/internal/version.Version=v" + version + \
    " -X phasionary/internal/version.Commit=" + commit + \
    " -X phasionary/internal/version.BuildDate=" + build_date

# Run from source, passing args through (e.g. `just run version`).
run *ARGS:
    go run -ldflags "{{ldflags}}" ./cmd/phasionary {{ARGS}}

# Run the sync server from source (`just server pair` prints a device code).
server *ARGS:
    go run -ldflags "{{ldflags}}" ./cmd/phasionary-server {{ARGS}}

# Serve the web app on :5173 with a sync server on :7777 behind it; the server embeds a build so its QR works.
web: _embed
    #!/usr/bin/env bash
    set -euo pipefail
    go build -ldflags "{{ldflags}}" -o bin/phasionary-server ./cmd/phasionary-server
    bin/phasionary-server &
    server=$!
    trap 'kill $server 2>/dev/null || true' EXIT
    # Lets the server print its pairing QR before vite's banner lands on it.
    sleep 3
    npm --prefix web run dev

# Build both binaries, with the web app embedded in the server.
build: _embed
    go build -ldflags "{{ldflags}}" -o phasionary ./cmd/phasionary
    go build -ldflags "{{ldflags}}" -o phasionary-server ./cmd/phasionary-server

_embed:
    [ -x web/node_modules/.bin/vite ] || npm --prefix web install
    npm --prefix web run build
    cp -r web/dist/. internal/webui/dist/

# Run every test; the nix builds type-check and unit-test the web app too.
test:
    go test ./...
    nix build
    nix build .#phasionary-server

# Format Go files and tidy go.mod.
fmt:
    gofmt -w cmd internal
    go mod tidy

# Bump version (X.Y.Z or major|minor|patch); updates VERSION, commits, tags.
bump version:
    ./scripts/bump.sh {{version}}
