# Phasionary - TUI Tech Stack

## Decision
The implementation will target Go + Bubble Tea.

## Shared Requirements
- Single binary distribution (Linux only for now)
- Local JSON files (one per project), no network dependency
- Config file + theme overrides
- JSON/CSV import/export
- Deterministic builds and reproducible releases

## Go + Bubble Tea
- TUI: bubbletea + lipgloss
- CLI: cobra + viper
- Tests: go test + testify

## Packaging
- Go: goreleaser
