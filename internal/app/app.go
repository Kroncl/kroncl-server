package app

import (
	"context"
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

	serverErrors := make(chan error, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.server.Run(ctx, serverErrors); err != nil {
			serverErrors <- err
		}
	}()

	// Ожидаем сигналов
	select {
	case err := <-serverErrors:
		log.Printf("Server error: %v", err)
		cancel()
	case sig := <-signals:
		log.Printf("Signal received: %v", sig)
		cancel()
	}

	// Graceful shutdown
	return a.shutdown()
}

func (a *Application) shutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Wait for goroutines
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

	// Close resources
	a.container.Close()

	log.Println("Application stopped")
	return nil
}
