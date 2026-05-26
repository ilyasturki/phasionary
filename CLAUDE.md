# CLAUDE.md

This file provides guidance to Codex CLI and Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Phasionary is a terminal-first, single-user project planning tool. Local-only, offline by default, keyboard-driven. The implementation targets Go with Bubble Tea for TUI. The app should look well on dark and light mode.

## Build Commands

```bash
go build -o phasionary ./cmd/phasionary
just test                          # Runs `go test ./...` then `nix build`
go test -v ./internal/domain/...   # Run tests for specific package
```

## Visual TUI Testing

Run `./testdata/vhs/run.sh [tape-name]` to verify UI changes. It rebuilds,
re-seeds an isolated data dir, runs each `.tape`, and writes PNG screenshots
to `/tmp/phas-vt/vhs-out/`. Read the PNGs, then `rm -rf /tmp/phas-vt/vhs-out`.
Add a new tape for any new mode, modal, or non-trivial keybinding.

