# No clipboard utility reachable: WAYLAND_DISPLAY is empty and XDG_RUNTIME_DIR
# points at a directory with no compositor socket in it, so the copy can only
# leave as OSC 52 and the status has to say so.
Output "/tmp/phas-vt/vhs-out/35_clipboard_fallback.gif"
Set Shell "bash"
Set Width 1200
Set Height 700
Set FontSize 16
Env PHASIONARY_DATA_PATH "/tmp/phas-vt/data"
Env PHASIONARY_CONFIG_PATH "/tmp/phas-vt/cfg"
Env XDG_RUNTIME_DIR "/tmp/phas-vt/no-wayland-socket"
Env WAYLAND_DISPLAY ""

Hide
Type "clear && ./phasionary -p Default"
Enter
Sleep 800ms
Show

Type "y"
Sleep 600ms
Screenshot "/tmp/phas-vt/vhs-out/35_clipboard_fallback.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms

Hide
Type "q"
