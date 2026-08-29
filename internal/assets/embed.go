// Package assets embeds the production frontend build so the application
// can ship as a single self-contained binary. When --web-dir is set the
// server serves that directory instead. The Vite build writes its output
// into the dist directory here (see web/vite.config.ts).
package assets

import "embed"

// DistFS contains the built single-page application.
//
//go:embed all:dist
var DistFS embed.FS
