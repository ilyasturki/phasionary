# Phasionary - TUI Design System

Terminal-first aesthetic: sharp edges, flat colors, and dense information. No gradients, no rounded corners, and no mouse dependency.

## Layout
- Header: app name, active project, active section, filter badges
- Left rail (toggleable): project list
- Main list: tasks grouped by category
- Details pane: selected task details + metadata
- Footer: key hints, status messages, and mode indicator

## Interaction Modes
- Normal mode: navigation and actions
- Edit mode: form fields and text entry
- Command mode: ":" command palette for jumps and batch actions

## Typography & Spacing
- Monospace only
- One-line rows for lists; wrap descriptions in details pane
- Consistent column widths for status, priority, deadlines

## Color Tokens (ANSI-first)

| Token | Meaning | ANSI Suggestion |
| --- | --- | --- |
| text-primary | Default text | Default terminal color |
| text-muted | Secondary text | Faint/gray |
| accent | Primary actions | Green |
| info | In progress | Blue |
| success | Completed | Green |
| warning | Due soon | Yellow |
| error | Overdue / destructive | Red |
| priority-high | High priority | Red |
| priority-medium | Medium priority | Yellow |
| priority-low | Low priority | Blue |

## Status & Priority Indicators
- Status labels are always visible (no color-only states)
- Priority uses label + color (e.g., "P:High")

## Accessibility & Fallbacks
- Must be usable with `--no-color`
- Provide text labels for icons and badges
- Keep contrast high across light/dark terminal themes
