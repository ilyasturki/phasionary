Output "/tmp/phas-vt/vhs-out/37_picker_add_delete.gif"
Set Shell "bash"
Set Width 1200
Set Height 700
Set FontSize 16
Env PHASIONARY_DATA_PATH "/tmp/phas-vt/data"
Env PHASIONARY_CONFIG_PATH "/tmp/phas-vt/cfg"

Hide
Type "clear && for n in Alpha Bravo Charlie; do ./phasionary project add $n -q >/dev/null 2>&1; done && ./phasionary -p Default"
Enter
Sleep 1500ms
Show

Ctrl+P
Sleep 500ms
Type "gj"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/37_picker_on_project.png"

Type "a"
Sleep 300ms
Type@40ms "Zulu"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/37_picker_add_from_row.png"

Escape
Sleep 400ms

Type "d"
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/37_picker_delete_confirm.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms

Type "n"
Sleep 300ms
Escape
Sleep 300ms

Hide
Type "q"
