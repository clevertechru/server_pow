package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/clevertechru/server_pow/internal/server"
	"github.com/clevertechru/server_pow/pkg/config"
)

func main() {
	configPath := flag.String("config", "config/server.yml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadServerConfig(*configPath)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		cfg = config.DefaultServerConfig()
	}

	handler, err := server.NewHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: handler,
	}

	log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
