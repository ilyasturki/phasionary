Output "/tmp/phas-vt/vhs-out/22_tags.gif"
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

# Cursor starts on the first task. `t` cycles the color; the dot stays its
# own color even though the row is focused (reverse).
Type "t"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/22_tag_green_focused.png"

# `T` opens the full tag editor: label field + color picker.
Type "T"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/22_editor_open.png"

# Type a label, switch to the color picker, choose magenta.
Type "urgent"
Sleep 200ms
Tab
Sleep 150ms
Type "jj"
Sleep 200ms
Screenshot "/tmp/phas-vt/vhs-out/22_editor_color.png"
Enter
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/22_row_labeled_focused.png"

# Move the cursor off the task so the inline dot + label show on an
# unfocused row too.
Type "j"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/22_row_labeled_unfocused.png"

# Info view still lists the tag.
Type "gi"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/22_tag_info.png"
Type "q"
Sleep 200ms

# Remove the tag via the editor: back onto the task, T, switch to color,
# navigate up to "None".
Type "k"
Sleep 200ms
Type "T"
Sleep 200ms
Tab
Sleep 150ms
Type "kkk"
Sleep 200ms
Screenshot "/tmp/phas-vt/vhs-out/22_editor_none.png"
Enter
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/22_row_removed.png"

# Bulk: visual-select two tasks and open the shared editor.
Type "v"
Sleep 150ms
Type "j"
Sleep 150ms
Type "T"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/22_bulk_editor.png"
Escape
Sleep 200ms

Hide
Type "q"
