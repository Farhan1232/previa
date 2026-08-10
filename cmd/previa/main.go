// Command previa runs the Previa frontend server.
//
// This milestone serves server-rendered HTML from a centralised in-memory mock
// data provider. Swapping in the production MySQL provider is a change to the
// data.Store construction in package app and nothing else.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"previa/pkg/app"
	"previa/pkg/config"
)

func main() {
	cfg := config.Load()

	handler, err := app.New(cfg)
	if err != nil {
		log.Fatalf("startup: %v", err)
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	go func() {
		log.Printf("Previa listening on http://localhost%s (%s)", cfg.Addr, app.Describe(cfg))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("Previa stopped")
}
