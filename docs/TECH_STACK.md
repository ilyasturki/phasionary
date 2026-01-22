# Phasionary - TUI Tech Stack

## Decision
The implementation will target Go + Bubble Tea.

## Shared Requirements
- Single binary distribution (Linux only for now)
- Local JSON data file, no network dependency
- Config file + theme overrides
- JSON/CSV import/export
- Deterministic builds and reproducible releases

## Go + Bubble Tea
- TUI: bubbletea + lipgloss
- CLI: cobra + viper
- DB: modernc.org/sqlite or mattn/go-sqlite3
- Migrations: goose or golang-migrate
- Tests: go test + testify

## Packaging
- Go: goreleaser
