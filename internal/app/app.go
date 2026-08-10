// Package app assembles the Previa application.
//
// Both entry points use it: cmd/previa runs it as a long-lived HTTP server,
// and api/index.go wraps it as a Vercel serverless function. Keeping the wiring
// in one place means the two deployments cannot drift apart.
package app

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"previa/internal/assets"
	"previa/internal/config"
	"previa/internal/data"
	"previa/internal/handlers"
	"previa/internal/view"
	"previa/web"
)

// New builds the fully wired HTTP handler.
func New(cfg config.Config) (http.Handler, error) {
	engine, err := view.New(templateFS(cfg), cfg.Dev)
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	// Swap this line for the MySQL-backed store when the backend lands; the
	// handlers depend only on the interfaces in package data.
	store := data.NewStore(time.Now())

	return handlers.New(store, engine, cfg).Routes(), nil
}

// templateFS returns the template filesystem.
//
// In dev mode templates are read from disk so edits show up on refresh.
// Otherwise the copy embedded in package web is used, which is what makes the
// binary self-contained on Vercel.
func templateFS(cfg config.Config) fs.FS {
	if cfg.Dev {
		if _, err := os.Stat(cfg.TemplateDir); err == nil {
			return os.DirFS(cfg.TemplateDir)
		}
	}
	sub, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		// Cannot happen: the directory is embedded at compile time.
		panic(err)
	}
	return sub
}

// Describe summarises the build for startup logging.
func Describe(cfg config.Config) string {
	return fmt.Sprintf("dev=%v maps=%v assets=[%s]",
		cfg.Dev, cfg.MapsKey != "", assets.Describe())
}
