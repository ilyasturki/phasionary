# Phasionary

Terminal-first, single-user project planning tool. Local-only, offline by default, keyboard-driven. This repository contains an ultra-minimal skeleton to validate the Go setup and project structure.

## Status

This is a starter scaffold: a tiny TUI placeholder plus minimal CLI commands for init, project list/add, and task list/add/status.

## Requirements

- Go 1.22+

## Build

```bash
go build -o phasionary ./cmd/phasionary
```

## Run

```bash
./phasionary           # launches the minimal TUI
./phasionary init      # creates the data directory and default project
```

## Data Location

- Default: `~/.local/share/phasionary/`
- Override: `PHASIONARY_DATA_PATH` or `--data <path>`

## CLI Examples

```bash
./phasionary project list
./phasionary project add "Client Portal"

./phasionary task add "Set up schema" --category Feature --priority high --deadline 2026-01-22 --estimate 2h
./phasionary task list --status todo --section current
./phasionary task status <id> completed
```

## Notes

- Task storage uses one JSON file per project.
- The TUI is intentionally minimal for now; it only confirms the app boots and the dependencies are wired.

## Docs

See `docs/` for product specs, data schema, keybindings, and target architecture.
