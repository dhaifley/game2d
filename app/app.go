package app

import "embed"

// FS is a file system containing the compiled app files.
//
//go:embed dist/*
var FS embed.FS
