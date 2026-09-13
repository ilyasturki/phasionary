package webui

import (
	"embed"
	"io/fs"
)

// dist/placeholder keeps the directory non-empty for go:embed; a build replaces it with the bundle.
//
//go:embed all:dist
var embedded embed.FS

var FS, _ = fs.Sub(embedded, "dist")
