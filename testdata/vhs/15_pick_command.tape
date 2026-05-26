Output "/tmp/phas-vt/vhs-out/15_pick_command.gif"
Set Shell "bash"
Set Width 1200
Set Height 700
Set FontSize 16
Env PHASIONARY_DATA_PATH "/tmp/phas-vt/data"
Env PHASIONARY_CONFIG_PATH "/tmp/phas-vt/cfg"

Hide
Type "clear && ./phasionary pick"
Enter
Sleep 1500ms
Show
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/15_pick_opens_picker.png"

Hide
Escape
Sleep 300ms
Type "q"
