// Package web embeds the compiled frontend assets into the Go binary.
// Run `npm run build` in the frontend directory (with outDir=../web/dist) before building the binary.
package web

import "embed"

//go:embed all:dist
var Dist embed.FS
