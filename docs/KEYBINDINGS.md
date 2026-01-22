# Phasionary - TUI Navigation & Keybindings

## Global Keys
- `q` Quit (with confirm if dirty)
- `?` Help / keybindings
- `/` Search (live filter)
- `:` Command palette
- `Tab` Next pane
- `Shift+Tab` Previous pane
- `Ctrl+r` Refresh from DB

## Navigation
- `j/k` or `↑/↓` Move selection
- `h/l` or `←/→` Collapse/expand category group
- `g/G` Top/bottom
- `PgUp/PgDn` Page navigation

## Actions
- `a` Add task
- `e` Edit selected
- `d` Delete selected
- `s` Change status
- `p` Change priority
- `m` Move section
- `c` Change category
- `n` New project
- `x` Switch project

## Modes
- Normal: navigation and shortcuts
- Edit: form fields, `Esc` to cancel, `Enter` to save
- Command: `:` prefix commands (e.g., `:project switch <name>`)

## Quick Capture
- `Ctrl+n` opens a minimal form (title + category)
- Optional fields via inline toggles: deadline, estimate, priority

## Filters
- Toggle filters with `f`
- Multi-select filters with `space`
- Clear filters with `Shift+f`

## Confirmations
- Destructive actions require `y/n`
- Bulk operations show a diff summary before apply
