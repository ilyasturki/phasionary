Output "/tmp/phas-vt/vhs-out/32_long_title.gif"
Set Shell "bash"
Set Width 1200
Set Height 700
Set FontSize 16
Env PHASIONARY_DATA_PATH "/tmp/phas-vt/data"
Env PHASIONARY_CONFIG_PATH "/tmp/phas-vt/cfg"

# Titles have no length limit. This tape covers what bounds one instead: the
# list clamps the row to ui.MaxLineRows, a search moves that window onto the
# match, the inline editor windows onto the cursor, and the info dialog scrolls
# so the whole title stays readable somewhere.
Hide
Type `clear && ./phasionary task add "$(for i in $(seq 1 150); do printf 'w%03d ' $i; done)END" -C Ergonomy -q`
Enter
Sleep 500ms
Type "clear && ./phasionary -p Default"
Enter
Sleep 1000ms
Show

Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/32_long_title_clamped.png"

# A match past the clamp still shows: the window opens on it.
Type "/w120"
Sleep 600ms
Enter
Sleep 800ms
Screenshot "/tmp/phas-vt/vhs-out/32_long_title_search_window.png"

Escape
Sleep 400ms

# The info dialog holds the title in full, and scrolls.
Type "gi"
Sleep 900ms
Screenshot "/tmp/phas-vt/vhs-out/32_long_title_info_top.png"

# A no-op keystroke, as above: the window is already at the top, so this
# redraws it unchanged and flushes the queued screenshot on the top state
# instead of on the frame G draws.
Type "k"
Sleep 400ms

Type "G"
Sleep 800ms
Screenshot "/tmp/phas-vt/vhs-out/32_long_title_info_bottom.png"

Type "q"
Sleep 600ms

# The inline editor stays one row, windowed onto the cursor at the end.
Enter
Sleep 1500ms
Screenshot "/tmp/phas-vt/vhs-out/32_long_title_edit_end.png"

# A no-op keystroke: the cursor is already at the end, so the row redraws
# unchanged. VHS writes a queued screenshot on the next rendered frame, and a
# still screen produces none — without this the capture above would land on the
# frame Ctrl+A draws instead.
Right
Sleep 400ms

Ctrl+A
Sleep 900ms
Screenshot "/tmp/phas-vt/vhs-out/32_long_title_edit_start.png"

Sleep 500ms
Escape
Sleep 400ms

Hide
Type "q"
Sleep 500ms
