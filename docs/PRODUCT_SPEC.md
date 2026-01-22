# Phasionary - Product Specification

## Overview

Phasionary is a terminal-first, single-user project planning tool. Local-only, offline by default, keyboard-driven.

**Goals:**
- Keyboard-first navigation
- Per-project JSON files in a data directory
- Fast task capture and triage
- Works over SSH and low-bandwidth terminals

**Non-goals (v1):** Multi-user, web/mobile clients, cloud sync, Gantt/Kanban/calendars.

## Entities

### Projects

Top-level container for tasks and categories.

| Attribute   | Required | Notes              |
|-------------|----------|--------------------|
| Name        | Yes      | Max 100 chars      |
| Description | No       | Optional summary   |

- Multiple projects allowed
- Tasks and categories belong to one project
- Default project created on `phasionary init`

### Categories

User-defined labels scoped to a project.

- Unique per project (case-insensitive)
- Tasks belong to one category
- Defaults on init: Feature, Fix, Ergonomy, Documentation, Research (Ergonomy is intentional shorthand)

### Status

Values: `todo`, `in_progress`, `completed`, `cancelled`

- Any status can transition to any other
- Completion date set on `completed`, cleared if reopened

### Sections

Values: `current`, `future`, `past`

- `past` requires `completed` or `cancelled` status
- Reopening moves task back to `current`

## Ordering

Tasks sort within each category by:

1. Priority (high > medium > low > none)
2. Deadline (earliest first, none last)
3. Time estimate (shortest first, none last)
4. Title (A-Z, case-insensitive)

## Workflows

- Launch into last-used project, current section
- Quick capture: title + category, optional fields inline
- Triage: change status, section, priority
- Category management with cascade delete (removes all nested tasks)
- Project switching via sidebar or command palette
- Search/filter by text, status, section, category, priority, overdue

## CLI

- Shares local JSON data directory with TUI
- Human-readable tables by default, JSON for scripting
- Focused on automation, import/export, quick edits

## Data

- Data directory auto-created on startup if missing; default project created via `phasionary init` or on first TUI launch
- Export/import: JSON and CSV
- Backups: JSON snapshots or exports
