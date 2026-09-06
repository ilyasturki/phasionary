Output "/tmp/phas-vt/vhs-out/05_options.gif"
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

Type ","
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/05_options_open.png"
# Idle renders no frame, so without this the capture lands after the `j` below.
Sleep 500ms

Type "j"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/05_options_priority_full.png"

Type " "
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/05_options_priority_icon.png"

Type "h"
Sleep 300ms
Type "l"
Sleep 300ms

Escape
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/05_options_tasks_icon_only.png"

Type ","
Sleep 300ms
Type "j"
Sleep 200ms
Type " "
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/05_options_priority_none.png"

Escape
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/05_options_tasks_no_color.png"

Type ","
Sleep 300ms
Type "j"
Sleep 200ms
Type " "
Sleep 300ms
Escape
Sleep 300ms

Type ","
Sleep 300ms
Type "jjj"
Sleep 200ms
Screenshot "/tmp/phas-vt/vhs-out/05_options_descriptions_off.png"
Type " "
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/05_options_descriptions_on.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms
Type " "
Sleep 200ms
Escape
Sleep 300ms

Hide
Type "q"
