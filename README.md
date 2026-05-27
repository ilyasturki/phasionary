# Phasionary

A terminal-first project planner for people who'd rather press `j` than reach for a mouse. Single binary, plain JSON on disk, no accounts, no network.

![Phasionary demo](assets/demo.gif)

## Install

### Arch Linux (AUR)

```bash
yay -S phasionary       # build from source
yay -S phasionary-bin   # prebuilt binary
```

### Nix

Run without installing:

```bash
nix run github:ilyasturki/phasionary
```

Or pin it in your system flake:

```nix
inputs.phasionary.url = "github:ilyasturki/phasionary";
# environment.systemPackages = [ inputs.phasionary.packages.x86_64-linux.default ];
```

### Prebuilt binary

Grab the latest `phasionary-linux-x64` or `phasionary-linux-arm64` from the [releases page](https://github.com/ilyasturki/phasionary/releases), `chmod +x`, drop in `$PATH`.

### From source

```bash
go build -o phasionary ./cmd/phasionary
```

## Quick start

```bash
phasionary                # Launch TUI — opens the last project, or the picker
phasionary -p "Life"      # Jump straight into a project by name
```

First run seeds a default project. Inside the TUI: `a` adds a task, `Space` cycles status, `i` opens task info, `?` shows every keybinding.

## CLI

Every TUI action has a CLI counterpart, useful for scripts and quick edits. All commands accept `-j` for JSON output.

```bash
phasionary tasks -s todo -C Feature                   # Filter tasks
phasionary task add -C Feature "Build widget"         # Capture
phasionary task status <id> in_progress               # Triage
phasionary export -f json -o project.json             # Snapshot
phasionary import project.md                          # Restore from markdown
```

Run `phasionary --help` for the full surface (projects, categories, config, completions).

## Configuration

Config lives at `~/.config/phasionary/config.json`. Override with `phasionary config set <key> <value>` or edit directly.

| Key | Values | Default | Description |
|-----|--------|---------|-------------|
| `status_display` | `text`, `icons` | `icons` | How task status renders in the TUI |
| `default_project` | project UUID | (none) | Project to open on launch |
| `priority_color` | `full`, `icon`, `none` | `full` | How priority is colored — full row, icon only, or off |

Paths can be relocated with `PHASIONARY_CONFIG_PATH` and `PHASIONARY_DATA_PATH` — both name a directory; `PHASIONARY_CONFIG_PATH` also accepts a path ending in `config.json` for convenience.

## Data

One project per JSON file under `~/.local/share/phasionary/projects/{uuid}.json`. UI state (fold state, last project per directory) sits next to it in `state.json`. Every change is written synchronously — no undo, but nothing is ever in-flight.

## License

[MIT](LICENSE)
