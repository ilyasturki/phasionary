Output "/tmp/phas-vt/vhs-out/36_help_columns.gif"
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

# The card opens on the first entry of the left column.
Type "?"
Sleep 600ms
Screenshot "/tmp/phas-vt/vhs-out/36_help_col_left.png"

# `l` crosses to Create & edit, landing on the nearest line.
Type "l"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/36_help_col_right.png"

# Down the right column, then `h` comes back to the nearest left-column entry.
# `delete` sits on the blank line between the two left groups, so the nearest
# entry is `search` above it — not `move up / down`, where the cursor started.
Type "jjj"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/36_help_col_right_bottom.png"
Type "h"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/36_help_col_back.png"

# `h` again has nowhere to go: the cursor holds still.
Type "h"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/36_help_col_no_wrap.png"

# The full reference is one column, so h/l is inert there.
Type "a"
Sleep 400ms
Type "jjl"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/36_help_col_reference_inert.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms

Hide
Type "?"
Sleep 200ms
Type "q"
