Output "/tmp/phas-vt/vhs-out/16_picker_scroll.gif"
Set Shell "bash"
Set Width 1200
Set Height 480
Set FontSize 16
Env PHASIONARY_DATA_PATH "/tmp/phas-vt/data"
Env PHASIONARY_CONFIG_PATH "/tmp/phas-vt/cfg"

# Seed enough projects that the list must scroll, then launch on Default. The
# run.sh harness re-seeds before every tape, so these extra projects stay
# isolated to this run.
Hide
Type "clear && for n in Alpha Bravo Charlie Delta Echo Foxtrot Golf Hotel India Juliet Kilo Lima Mike November Oscar; do ./phasionary project add $n -q >/dev/null 2>&1; done && ./phasionary -p Default"
Enter
Sleep 2500ms
Show

# Open the picker: "+ New Project" is pinned at the top, the window sizes
# itself to the terminal, and each project shows its task progress.
Ctrl+P
Sleep 500ms
Screenshot "/tmp/phas-vt/vhs-out/16_picker_open.png"

# ctrl+d pages down half a screen; the pinned New Project stays put while the
# project list scrolls (note the "more above" indicator).
Ctrl+D
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/16_picker_scrolled.png"

# The pinned New Project affordance is reached by going to the first project
# (g) and stepping up past it (k).
Type "g"
Sleep 200ms
Type "k"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/16_picker_new_project.png"

Hide
Escape
Sleep 300ms
Type "q"
