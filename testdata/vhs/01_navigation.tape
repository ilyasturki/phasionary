Output "/tmp/phas-vt/vhs-out/01_navigation.gif"
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

Type "jjjjj"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/01_after_jjjjj.png"

Type "G"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/01_after_G.png"

Type "gg"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/01_after_gg.png"

Type "}"
Sleep 300ms
Type "}"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/01_after_next_category.png"

Hide
Type "q"
