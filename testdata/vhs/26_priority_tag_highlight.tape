Output "/tmp/phas-vt/vhs-out/26_priority_tag_highlight.gif"
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

# Cursor starts on the first task "Build the main dashboard" (high priority ▲).
# On a highlighted row the priority icon must keep its color as a reverse
# color block instead of vanishing into the monochrome selection bar.
Screenshot "/tmp/phas-vt/vhs-out/26_priority_highlight.png"

# Color-only tag (green): the tag block and the priority block sit side by
# side, each in its own color.
Type "t"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/26_coloronly_tag_priority.png"

# Labeled tag beside a priority icon: the separator space between the tag
# block and the priority block is filled with the tag color so the two
# colored blocks connect (no monochrome gap).
Type "T"
Sleep 300ms
Type "urgent"
Sleep 200ms
Tab
Sleep 150ms
Type "jj"
Sleep 200ms
Enter
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/26_labeled_tag_priority.png"

# Regression: a labeled tag on a medium task (no priority icon) still keeps
# its trailing separator in the selection band so the block hugs the label.
Type "j"
Sleep 200ms
Type "T"
Sleep 300ms
Type "later"
Sleep 200ms
Tab
Sleep 150ms
Type "j"
Sleep 200ms
Enter
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/26_labeled_tag_no_priority.png"

Hide
Type "q"
