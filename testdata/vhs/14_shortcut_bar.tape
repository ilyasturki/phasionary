Output "/tmp/phas-vt/vhs-out/14_shortcut_bar.gif"
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

Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/14_shortcut_bar_default.png"

Type "v"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/14_shortcut_bar_visual.png"

Escape
Sleep 300ms

Type ","
Sleep 300ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Screenshot "/tmp/phas-vt/vhs-out/14_shortcut_bar_options_focused.png"

Type " "
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/14_shortcut_bar_options_off.png"

Escape
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/14_shortcut_bar_hidden.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms

Type ","
Sleep 300ms
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type " "
Sleep 300ms
Escape
Sleep 300ms

Hide
Type "q"
