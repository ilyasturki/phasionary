Output "/tmp/phas-vt/vhs-out/02_help.gif"
Set Shell "bash"
Set Width 1200
Set Height 700
Set FontSize 16
Env PHASIONARY_DATA_PATH "/tmp/phas-vt/data"
Env PHASIONARY_CONFIG_PATH "/tmp/phas-vt/cfg"

Hide
Type "clear && ./phasionary -p Default"
Enter
Sleep 800ms
Show

Type "?"
Sleep 600ms
Screenshot "/tmp/phas-vt/vhs-out/02_help_open.png"

# ? does NOT close help (only q/esc do) — see app.go:103
Type "?"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/02_help_still_open.png"

Hide
Type "q"
