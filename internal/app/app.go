package app

import (
	"context"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/di"
	"kroncl-server/internal/server"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Application struct {
	container *di.Container
	server    *server.Server
	wg        sync.WaitGroup
}

func New(cfg *config.Config) (*Application, error) {
	ctx := context.Background()

	container, err := di.NewContainer(ctx, cfg)
	if err != nil {
		return nil, err
	}

	srv := server.New(cfg, container)

	return &Application{
		container: container,
		server:    srv,
	}, nil
}

func (a *Application) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start-core-workers [metrics]
	if err := a.container.CoreWorkers.Start(); err != nil {
		return fmt.Errorf("failed to start metrics worker: %w", err)
	}

	serverErrors := make(chan error, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// start-http
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.server.Run(ctx, serverErrors); err != nil {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		log.Printf("Server error: %v", err)
		cancel()
	case sig := <-signals:
		log.Printf("Signal received: %v", sig)
		cancel()
	}

	return a.shutdown()
}

func (a *Application) shutdown() error {
	log.Println("Shutting down...")

	// stop-core-workers
	if a.container.CoreWorkers != nil {
		a.container.CoreWorkers.Stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// stop-http
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout")
	case <-done:
		log.Println("All goroutines stopped")
	}

	if a.container.StorageService != nil {
		a.container.StorageService.CloseAll()
	}

	a.container.Close()

	log.Println("Application stopped")
	return nil
}
