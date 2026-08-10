// Package handler is the Vercel entry point.
//
// Vercel's Go runtime invokes an exported Handler per request rather than
// running a long-lived server, so the application is built once on the first
// request and reused for the lifetime of the warm instance.
//
// Static assets never reach this function: files under public/ are served
// straight from Vercel's CDN, and vercel.json only rewrites what the
// filesystem does not resolve.
package handler

import (
	"net/http"
	"sync"

	"previa/pkg/app"
	"previa/pkg/config"
)

var (
	once    sync.Once
	router  http.Handler
	initErr error
)

// Handler serves every dynamic request.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		cfg := config.Load()
		// Templates are embedded, so dev-mode disk reloading must stay off
		// here regardless of the environment.
		cfg.Dev = false
		router, initErr = app.New(cfg)
	})

	if initErr != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	router.ServeHTTP(w, r)
}
