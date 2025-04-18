package main

import (
	"flag"
	"log"

	"github.com/clevertechru/server_pow/internal/server"
	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/quotes"
)

func main() {
	serverConfigPath := flag.String("config", "config/server.yml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadServerConfig(*serverConfigPath)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		cfg = config.DefaultServerConfig()
	}

	// Initialize quotes storage
	if err := quotes.Init(cfg.Server.Quotes.File); err != nil {
		log.Fatalf("Failed to initialize quotes storage: %v", err)
	}

	httpServer, err := server.NewHTTPServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

	// Start server based on protocol
	if cfg.Server.Protocol == "https" {
		if cfg.Server.TLS.CertFile == "" || cfg.Server.TLS.KeyFile == "" {
			log.Fatal("TLS certificate and key files are required for HTTPS")
		}
		if err := httpServer.StartTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile); err != nil {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	} else {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}
}
