Output "/tmp/phas-vt/vhs-out/23_tag_copy.gif"
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

# Give the first task a tag: magenta + "backend".
Type "T"
Sleep 250ms
Type "backend"
Sleep 150ms
Tab
Sleep 150ms
Type "jj"
Sleep 150ms
Enter
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/23_source_tagged.png"

# Copy the tag (gt): stashes it for painting and puts the label on the system
# clipboard. The status line confirms what was copied.
Type "gt"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/23_after_copy.png"

# Move down and paste the tag onto another task (p, since the tag was copied
# last) — title/status untouched.
Type "j"
Sleep 200ms
Type "p"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/23_after_paste.png"

# Paint the tag onto a visual selection with p.
Type "j"
Sleep 150ms
Type "v"
Sleep 150ms
Type "j"
Sleep 150ms
Type "p"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/23_after_visual_paste.png"

# gy on a labeled task exposes the tag label as a copy-to-clipboard candidate.
Type "gy"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/23_yank_menu.png"
Escape
Sleep 200ms

Hide
Type "q"
