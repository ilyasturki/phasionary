# Phasionary — GitHub social preview screenshot
# Usage: vhs assets/social-preview.tape
# Produces a 2:1 PNG suitable for GitHub's Social preview setting.

Output assets/social-preview.gif
Set FontSize 22
Set FontFamily "IosevkaTerm Nerd Font"
Set Width 1280
Set Height 640
Set Padding 24
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

# Highlight a high-priority task with priority arrow visible
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "}"
Sleep 200ms
Type "j"
Sleep 200ms
Type "j"
Sleep 700ms

Screenshot assets/social-preview.png
Sleep 500ms

Hide
Type "q"
