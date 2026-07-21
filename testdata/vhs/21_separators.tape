Output "/tmp/phas-vt/vhs-out/21_separators.gif"
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

# Move onto the first task of the first category.
Type "j"
Sleep 300ms

# Insert a bare separator below the cursor (instant, no editor).
Type "-"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/21_bare_separator.png"

# Give it a label with Enter, type, commit.
Enter
Sleep 300ms
Type "Milestone v1"
Sleep 300ms
Enter
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/21_labeled_separator.png"

# Reorder it down past the next task.
Type "J"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/21_separator_moved.png"

# Delete it instantly (no confirm dialog).
Type "d"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/21_separator_deleted.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms

Hide
Type "q"
Enter
Sleep 300ms
