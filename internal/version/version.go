// Package version holds build-time version metadata embedded into the binary.
package version

// Version is the Athenaeum server release (set via -ldflags at build time).
var Version = "dev"

// WebVersion is the embedded frontend version from web/package.json.
var WebVersion = "dev"
