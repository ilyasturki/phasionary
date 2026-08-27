Output "/tmp/phas-vt/vhs-out/32_description_preview.gif"
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

# Expand inline descriptions: the long dashboard description shows a 3-line
# preview with a "… +N more lines" tail while the cursor is elsewhere.
Type "zd"
Sleep 800ms
Screenshot "/tmp/phas-vt/vhs-out/32_preview_truncated.png"

# Move the cursor onto the description row: it expands to the full body.
Type "j"
Sleep 800ms
Screenshot "/tmp/phas-vt/vhs-out/32_preview_focused_full.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms

Hide
Type "q"
Enter
Sleep 300ms
