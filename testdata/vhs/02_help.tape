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

Type "jjjjj"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/02_help_focus_moved.png"

Ctrl+D
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/02_help_scrolled.png"

Ctrl+U
Sleep 300ms
Type "k"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/02_help_focus_back.png"

Type "?"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/02_help_closed.png"

Type "?"
Sleep 400ms
Type "jj"
Sleep 200ms
Enter
Sleep 600ms
Screenshot "/tmp/phas-vt/vhs-out/02_help_enter_dispatches.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms

Hide
Type "q"
