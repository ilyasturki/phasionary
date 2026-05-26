Output "/tmp/phas-vt/vhs-out/13_status_cycle.gif"
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

# Move onto the second task ("Add user preferences panel" — starts as todo)
Type "j"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/13_initial.png"

# Cycle forward four times: todo -> in_progress -> completed -> cancelled -> todo
Type " "
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/13_fwd_1_in_progress.png"
Type " "
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/13_fwd_2_completed.png"
Type " "
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/13_fwd_3_cancelled.png"
Type " "
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/13_fwd_4_todo.png"

# shift+space (reverse cycle) requires kitty keyboard protocol; covered by unit
# tests in task_test.go since VHS 0.x cannot synthesize the chord.

Hide
Type "q"
