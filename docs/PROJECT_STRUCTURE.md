# Phasionary - Project Structure (TUI)

Target structure for Go (Bubble Tea).

## Go (Bubble Tea)

```
phasionary/
├── cmd/phasionary/
│   └── main.go
├── internal/
│   ├── app/          # App state + update loop
│   ├── ui/           # Views, components, styles
│   ├── cli/          # Cobra commands
│   ├── domain/       # Entities + sorting rules
│   ├── data/         # JSON storage access
│   ├── config/       # Config loading
│   └── export/       # Import/export helpers
├── migrations/
├── docs/
```

