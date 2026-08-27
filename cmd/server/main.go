package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"thermal-cycle-lab/internal/api"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/engine"
	"thermal-cycle-lab/internal/store"
	"time"
)

func main() {
	repository := store.NewMemory()
	source := clock.System{}
	catalogService := catalog.New(repository, source)
	controller := engine.New(repository, source)
	cancelObserver := controller.Observe(func(event engine.Event) { log.Printf("experiment event run=%s kind=%s", event.RunID, event.Kind) })
	defer cancelObserver()
	server := api.New(catalogService, controller)
	address := envOr("THERMAL_LAB_ADDR", ":8096")
	httpServer := &http.Server{Addr: address, Handler: server.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}
	stopped := make(chan struct{})
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals
		ctx, cancel := signalContext(8 * time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
		close(stopped)
	}()
	log.Printf("thermal cycle laboratory listening on %s", address)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve failed: %v", err)
	}
	select {
	case <-stopped:
	case <-time.After(9 * time.Second):
	}
}
func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
