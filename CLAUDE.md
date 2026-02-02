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
  app/                    # App state + Bubble Tea update loop
    components/           # Reusable UI components (Modal, TaskLineRenderer)
    modes/                # UI mode state machine
    selection/            # Navigation/selection manager
  ui/                     # Lipgloss styles
  cli/                    # Cobra commands
  domain/                 # Entities (Project, Category, Task) + sorting rules
  data/                   # JSON storage access
  config/                 # Config loading
  export/                 # Import/export helpers (JSON, CSV)
```

## Domain Model

- **Projects**: Top-level containers, stored as individual JSON files (`{uuid}.json`)
- **Categories**: User-defined labels scoped to a project (Feature, Fix, Ergonomy, Documentation, Research)
- **Tasks**: Belong to one category; have status (`todo`, `in_progress`, `completed`, `cancelled`) and section (`current`, `future`, `past`)

Task sort order within categories: Priority > Deadline > Time estimate > Title (A-Z)

## TUI Modes

- **Normal**: Navigation and actions
- **Edit**: Form fields and text entry
- **Help**: Help dialog overlay
- **ConfirmDelete**: Deletion confirmation
- **Options**: Settings dialog
- **ProjectPicker**: Project switching

## Key Constraints

- Single binary distribution (Linux only for v1)
- No network dependency
- Must work over SSH and low-bandwidth terminals
- All timestamps in UTC ISO 8601
- Project/category names are case-insensitive unique

## Key Architectural Decisions

- **State separation**: Model holds domain data, UI state, and dependencies as separate concerns
- **Persistence**: Synchronous full-project save on every change; no undo support
- **Navigation**: Flat position list derived from nested structure; rebuilt after structural changes
- **Config**: `~/.config/phasionary/config.json`

## Code Style Requirements

- Write comments only when necessary, prefer self-documenting code with clear naming
