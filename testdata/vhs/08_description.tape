Output "/tmp/phas-vt/vhs-out/08_description.gif"
Set Shell "bash"
Set Width 1200
Set Height 700
Set FontSize 16
Env PHASIONARY_DATA_PATH "/tmp/phas-vt/data"
Env PHASIONARY_CONFIG_PATH "/tmp/phas-vt/cfg"

Hide
Type "clear && ./phasionary -p Default"
Enter
Sleep 1000ms
Show

Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/08_description_indicator.png"

Sleep 200ms
Type "i"
Sleep 800ms
Screenshot "/tmp/phas-vt/vhs-out/08_description_info.png"

Type "q"
Sleep 300ms

Hide
Type "q"
