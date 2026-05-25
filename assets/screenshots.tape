# Phasionary — feature screenshots for the README
# Usage: vhs assets/screenshots.tape
# Produces task-list.png and task-info.png in assets/.

Output assets/screenshots.gif
Set FontSize 28
Set FontFamily "IosevkaTerm Nerd Font"
Set Width 2400
Set Height 1400
Set Padding 40
Set Framerate 30
Set Theme { "name": "Adwaita Dark", "background": "#1d1d20", "foreground": "#ffffff", "cursor": "#ffffff", "selection": "#ffffff", "black": "#241f31", "red": "#c01c28", "green": "#2ec27e", "yellow": "#f5c211", "blue": "#1e78e4", "magenta": "#9841bb", "cyan": "#0ab9dc", "white": "#c0bfbc", "brightBlack": "#5e5c64", "brightRed": "#ed333b", "brightGreen": "#57e389", "brightYellow": "#f8e45c", "brightBlue": "#51a1ff", "brightMagenta": "#c061cb", "brightCyan": "#4fd2fd", "brightWhite": "#f6f5f4" }
Set TypingSpeed 50ms
Set Shell bash

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

Type "./phasionary -p Life"
Enter
Sleep 1.5s

# --- task-list.png: main view, cursor on a high-priority task ---
Type "j"
Sleep 150ms
Type "j"
Sleep 150ms
Type "}"
Sleep 150ms
Type "j"
Sleep 150ms
Type "j"
Sleep 600ms

Screenshot assets/task-list.png
Sleep 500ms

# --- task-info.png: rich task info modal with description ---
Type "i"
Sleep 1s

Screenshot assets/task-info.png
Sleep 500ms

Escape
Sleep 400ms

Hide
Type "q"
