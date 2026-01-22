# Phasionary - CLI Commands

## Overview
The CLI is a thin automation layer over the local JSON data directory. The TUI is the default experience.

## Global Options
- `--data <path>` override data directory path
- `--project <name|id>` target project for commands

## Commands

### Setup
- `phasionary` or `phasionary tui` - Launch TUI (default)
- `phasionary init` - Create data directory with default project and categories (running TUI without prior init will auto-create defaults)

### Projects
- `phasionary project list`
- `phasionary project add <name>`

### Tasks
- `phasionary task list [--status todo|in_progress|completed|cancelled] [--section current|future|past]`
- `phasionary task add <title> --category <name> [--priority high|medium|low] [--deadline <YYYY-MM-DD>] [--estimate <number><unit>]`
  - `--deadline`: ISO 8601 date format (e.g., `2026-01-22`)
  - `--estimate`: number followed by unit `m` (minutes), `h` (hours), or `d` (days). Examples: `30m`, `2h`, `1d`
- `phasionary task status <id> <todo|in_progress|completed|cancelled>`
