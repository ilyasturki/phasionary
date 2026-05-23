Output "/tmp/phas-vt/vhs-out/03_add_task.gif"
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

# Adds under the currently-focused category
Type "a"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/03_add_prompt.png"

Type@40ms "Visual regression test pass"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/03_add_typed.png"

Enter
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/03_add_saved.png"

Hide
Type "q"
