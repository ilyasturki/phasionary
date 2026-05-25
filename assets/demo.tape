# Phasionary — README hero demo
# Usage: vhs assets/demo.tape
#
# Requires `vhs` and a built binary in the repo root (`go build ./cmd/phasionary`).
# The tape seeds an isolated data dir at /tmp/phas-demo/, imports the demo
# project ("Life"), then drives the app through its main features.

Output assets/demo.gif
Set FontSize 28
Set FontFamily "IosevkaTerm Nerd Font"
Set Width 2400
Set Height 1400
Set Padding 40
Set Framerate 30
Set Theme { "name": "Adwaita Dark", "background": "#1d1d20", "foreground": "#ffffff", "cursor": "#ffffff", "selection": "#ffffff", "black": "#241f31", "red": "#c01c28", "green": "#2ec27e", "yellow": "#f5c211", "blue": "#1e78e4", "magenta": "#9841bb", "cyan": "#0ab9dc", "white": "#c0bfbc", "brightBlack": "#5e5c64", "brightRed": "#ed333b", "brightGreen": "#57e389", "brightYellow": "#f8e45c", "brightBlue": "#51a1ff", "brightMagenta": "#c061cb", "brightCyan": "#4fd2fd", "brightWhite": "#f6f5f4" }
Set TypingSpeed 45ms
Set Shell bash

# --- Setup: isolated data dir + seed import (hidden) ---
Hide
Type "export PHASIONARY_DATA_PATH=/tmp/phas-demo/data"
Enter
Type "export PHASIONARY_CONFIG_PATH=/tmp/phas-demo/cfg"
Enter
Type "rm -rf /tmp/phas-demo && mkdir -p /tmp/phas-demo/data /tmp/phas-demo/cfg"
Enter
Type "./phasionary import assets/demo.seed.json -q"
Enter
Type "clear"
Enter
Show

# Launch — pause so the viewer can read the layout
Type "./phasionary -p Life"
Enter
Sleep 2.5s

# --- Tour: move through Home ---
Type "j"
Sleep 350ms
Type "j"
Sleep 350ms
Type "j"
Sleep 600ms

# Jump to the next category (Health)
Type "}"
Sleep 700ms

# Land on the "Run 5 km" task
Type "j"
Sleep 250ms
Type "j"
Sleep 500ms

# --- Task info: show description with notes & context ---
Type "i"
Sleep 2.2s
Escape
Sleep 500ms

# --- Toggle status: cycle todo/in_progress/completed/cancelled ---
Type "j"
Sleep 350ms
Space
Sleep 700ms

# --- Change priority on the focused task ---
Type "l"
Sleep 600ms

# --- Add a new task in the current category ---
Type "a"
Sleep 500ms
Type@55ms "Plan a long weekend trip in June"
Sleep 400ms
Enter
Sleep 700ms

# --- Set a time estimate via the picker ---
Type "t"
Sleep 600ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 400ms
Enter
Sleep 700ms

# --- Filter: hub → Status → toggle completed off ---
Type "f"
Sleep 800ms
Enter
Sleep 600ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Space
Sleep 700ms
Escape
Sleep 400ms
Escape
Sleep 800ms

# Re-open the filter and clear it so all tasks come back
Type "f"
Sleep 500ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 300ms
Enter
Sleep 600ms
Escape
Sleep 500ms

# --- Jump to Learning and open URLs on a task with two links ---
Type "}"
Sleep 500ms
Type "j"
Sleep 250ms
Type "j"
Sleep 250ms
Type "j"
Sleep 700ms

Type "gx"
Sleep 2s
Type "j"
Sleep 800ms
Escape
Sleep 700ms

# --- Visual select: extend a range and copy as markdown checklist ---
Type "v"
Sleep 700ms
Type "j"
Sleep 300ms
Type "j"
Sleep 300ms
Type "j"
Sleep 800ms
Type "Y"
Sleep 1.2s
Escape
Sleep 500ms

# --- Fold everything, then unfold ---
Type "zc"
Sleep 1.5s
Type "zo"
Sleep 1.2s

# --- Help dialog: show the full keymap ---
Type "?"
Sleep 2s
Type "j"
Sleep 220ms
Type "j"
Sleep 220ms
Type "j"
Sleep 220ms
Type "j"
Sleep 220ms
Type "j"
Sleep 220ms
Type "j"
Sleep 800ms
Escape
Sleep 800ms

# --- Quit ---
Type "q"
Sleep 800ms
