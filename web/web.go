// Package web embeds the HTML templates into the binary.
//
// Templates are compiled in rather than read from disk so the application runs
// identically as a long-lived server and as a Vercel serverless function,
// where only the Go build output ships and there is no repository beside it.
package web

import "embed"

// Templates holds every file under web/templates.
//
// The all: prefix keeps files that begin with "_" or "." — none exist today,
// but a partial named _foo.html would otherwise be silently dropped.
//
//go:embed all:templates
var Templates embed.FS
