Output "/tmp/phas-vt/vhs-out/18_reverse_categories.gif"
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

# Seed order top-to-bottom: Feature, Fix, Ergonomy, Documentation, Research.
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/18_before_reverse.png"

# S flips the whole category order in one shot: Research now on top, Feature last.
Type "S"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/18_after_reverse.png"

# S is its own inverse: pressing it again restores the original order.
Type "S"
Sleep 400ms
Screenshot "/tmp/phas-vt/vhs-out/18_toggled_back.png"

Hide
Type "q"
