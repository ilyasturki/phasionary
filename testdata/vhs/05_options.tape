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

Type "o"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/05_options_open.png"

Escape
Sleep 300ms

Hide
Type "q"
