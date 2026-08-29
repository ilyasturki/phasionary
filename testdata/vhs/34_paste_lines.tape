Output "/tmp/phas-vt/vhs-out/34_paste_lines.gif"
Set Shell "bash"
Set Width 1200
Set Height 700
Set FontSize 16
Env PHASIONARY_DATA_PATH "/tmp/phas-vt/data"
Env PHASIONARY_CONFIG_PATH "/tmp/phas-vt/cfg"

# Seeds the SYSTEM clipboard, which `p` reads once the internal one is empty.
# Heads up: this overwrites your real clipboard. VHS cannot send a bracketed
# paste, so the add-editor path is covered by paste_lines_test.go instead.
Hide
Type "clear && printf '## Sprint 12\nDesign the API\n- [ ] write handlers\n- [~] review\n- [x] write tests\nShip it\n' | (wl-copy 2>/dev/null || xclip -selection clipboard 2>/dev/null || xsel --clipboard --input)"
Enter
Sleep 500ms
Type "clear && ./phasionary -p Default"
Enter
Sleep 800ms
Show

Type "j"
Sleep 300ms
Screenshot "/tmp/phas-vt/vhs-out/34_before_paste.png"

# The bare lines sit beside bulleted ones, so they become the separators that
# group them; [~] keeps its in-progress status.
Type "p"
Sleep 700ms
Screenshot "/tmp/phas-vt/vhs-out/34_pasted.png"
# Let the final capture land before quitting (see run.sh).
Sleep 500ms

Hide
Type "q"
Enter
Sleep 300ms
