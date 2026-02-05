package main

import (
	"kroncl-server/internal/app"
	"kroncl-server/internal/config"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal("Failed to create application:", err)
	}

	if err := application.Run(); err != nil {
		log.Fatal("Application error:", err)
	}
}
