package static

import "embed"

// FS is a file system containing static files.
//
//go:embed * all:scripts
var FS embed.FS
