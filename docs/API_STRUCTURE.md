# Phasionary - CLI Commands

## Overview
The CLI is a thin automation layer over the local JSON data file. The TUI is the default experience.

## Global Options
- `--data <path>` override data file path
- `--project <name|id>` target project for commands

## Commands

### Setup
- `phasionary` or `phasionary tui` - Launch TUI (default)
- `phasionary init` - Create data file with default project and categories

### Projects
- `phasionary project list`
- `phasionary project add <name>`

### Tasks
- `phasionary task list [--status todo|in_progress|completed|cancelled] [--section current|future|past]`
- `phasionary task add <title> --category <name> [--priority high|medium|low]`
- `phasionary task status <id> <todo|in_progress|completed|cancelled>`
