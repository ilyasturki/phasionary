# CLAUDE.md

This file provides guidance to Codex CLI and Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Phasionary is a terminal-first, single-user project planning tool. Local-only, offline by default, keyboard-driven. The implementation targets Go with Bubble Tea for TUI.

## Tech Stack

- **TUI**: bubbletea + lipgloss
- **CLI**: cobra + viper
- **Storage**: Local JSON files (one per project) in `~/.local/share/phasionary/`
- **Tests**: go test + testify
- **Packaging**: goreleaser

## Build Commands

```bash
go build -o phasionary ./cmd/phasionary
go test ./...
go test -v ./internal/domain/...  # Run tests for specific package
```

## Architecture

```
cmd/phasionary/main.go    # Entry point
internal/
  app/      # App state + Bubble Tea update loop
  ui/       # Views, components, lipgloss styles
  cli/      # Cobra commands
  domain/   # Entities (Project, Category, Task) + sorting rules
  data/     # JSON storage access
  config/   # Config loading
  export/   # Import/export helpers (JSON, CSV)
```

## Domain Model

- **Projects**: Top-level containers, stored as individual JSON files (`{uuid}.json`)
- **Categories**: User-defined labels scoped to a project (Feature, Fix, Ergonomy, Documentation, Research)
- **Tasks**: Belong to one category; have status (`todo`, `in_progress`, `completed`, `cancelled`) and section (`current`, `future`, `past`)

Task sort order within categories: Priority > Deadline > Time estimate > Title (A-Z)

## TUI Modes

- **Normal**: Navigation and actions
- **Edit**: Form fields and text entry
- **Command**: `:` prefix for palette commands

## Key Constraints

- Single binary distribution (Linux only for v1)
- No network dependency
- Must work over SSH and low-bandwidth terminals
- All timestamps in UTC ISO 8601
- Project/category names are case-insensitive unique

## Code Style Requirements

- Write comments only when necessary, prefer self-documenting code with clear naming
