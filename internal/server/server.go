package server

import (
	"context"
	"kroncl-server/internal/config"
	"kroncl-server/internal/di"
	"kroncl-server/internal/router"
	"log"
	"net/http"
)

type Server struct {
	config    *config.Config
	container *di.Container
	http      *http.Server
}

func New(cfg *config.Config, container *di.Container) *Server {
	return &Server{
		config:    cfg,
		container: container,
	}
}

func (s *Server) Run(ctx context.Context, errors chan<- error) error {
	r := router.New(s.config, s.container)

	s.http = &http.Server{
		Addr:         s.config.Server.Host + ":" + s.config.Server.Port,
		Handler:      r,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	go func() {
		log.Printf("🚀 Server started on http://%s", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errors <- err
		}
	}()

	<-ctx.Done()
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}
