Output "/tmp/phas-vt/vhs-out/30_visual_status.gif"
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

# Give the block mixed statuses first, so the shared-step behavior is visible:
# advance the second task one step past the first.
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Space
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/30_mixed_statuses.png"

# Anchor on the first task and extend down across the mixed block.
Type "k"
Sleep 200ms
Type "v"
Sleep 300ms
Type "j"
Sleep 200ms
Type "j"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/30_range_selected.png"

# Space steps the whole range from the FIRST task's status — the mixed block
# lands on one shared status.
Space
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/30_status_unified.png"

# Repeated presses walk the block on together; visual mode stays active.
Space
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/30_status_advanced.png"

# (Shift+Space walks the block back one step; VHS can't encode that keystroke,
# so the reverse direction is covered by TestVisualCycleStatus_ReverseWalksBackwards.)

Hide
Escape
Sleep 200ms
Type "q"
Enter
Sleep 300ms
