Output "/tmp/phas-vt/vhs-out/25_help_filter.gif"
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
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/25_help_open.png"

Type "/"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/25_help_filter_open.png"

Type "add"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/25_help_filter_by_label.png"

Escape
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/25_help_filter_cleared.png"

Type "/"
Sleep 200ms
Type "markdown"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/25_help_filter_display_only.png"

Escape
Sleep 200ms
Type "/"
Sleep 200ms
Type "opt"
Sleep 400ms
Enter
Sleep 600ms
Screenshot "/tmp/phas-vt/vhs-out/25_help_filter_enter_runs.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms

Hide
Type "q"
