#!/usr/bin/env bash
# Run VHS tapes against an isolated Phasionary data dir.
#
# Usage: ./testdata/vhs/run.sh [tape-name ...]
#   ./testdata/vhs/run.sh                  # run all
#   ./testdata/vhs/run.sh 03_add_task      # run one
set -euo pipefail

root=$(git rev-parse --show-toplevel 2>/dev/null) || {
    echo "run.sh must be invoked from within the phasionary git repo" >&2
    exit 1
}
cd "$root"

DATA="/tmp/phas-vt/data"
CFG="/tmp/phas-vt/cfg"
OUT="/tmp/phas-vt/vhs-out"

# Tapes that mutate state need a fresh seed before they run.
MUTATING_TAPES=(03_add_task 04_toggle_and_filter 05_options 06_reorder_across_categories 07_open_url 09_visual_select 10_reload 12_undo_redo 13_status_cycle)

rm -rf "$OUT" && mkdir -p "$OUT"
go build -o phasionary ./cmd/phasionary

mapfile -t TAPES < <(
    if [ "$#" -gt 0 ]; then
        for arg in "$@"; do
            name=$(basename "${arg%.tape}")
            echo "testdata/vhs/${name}.tape"
        done
    else
        find testdata/vhs -maxdepth 1 -name '*.tape' | sort
    fi
)

seed() { BIN="$(pwd)/phasionary" ./testdata/vhs/seed.sh "$DATA" "$CFG" >/dev/null; }
is_mutating() {
    local t=$1 m
    for m in "${MUTATING_TAPES[@]}"; do [[ $t == "$m" ]] && return 0; done
    return 1
}

dirty=1
for tape in "${TAPES[@]}"; do
    echo
    echo "=== $(basename "$tape") ==="
    name=$(basename "$tape" .tape)
    (( dirty )) && seed && dirty=0
    vhs "$tape"
    is_mutating "$name" && dirty=1
done

echo
echo "Screenshots in $OUT (delete after verification):"
find "$OUT" -name '*.png' | sort
