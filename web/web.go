package web

import "embed"

// Assets holds embedded frontend web UI files.
//go:embed index.html
var Assets embed.FS
