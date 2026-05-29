Output "/tmp/phas-vt/vhs-out/08_project_picker.gif"
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

# Open the project picker
Ctrl+P
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/08_picker_open.png"

# Navigate down to the "+ New Project" affordance
Type "j"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/08_picker_on_add.png"

# Enter add mode and type a name to confirm the dialog width is stable
Enter
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/08_picker_add_empty.png"

Type@40ms "Side Project"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/08_picker_add_typed.png"

# Cancel without creating so the seed stays intact
Escape
Sleep 300ms
Escape
Sleep 300ms

Hide
Type "q"
