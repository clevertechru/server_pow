package main

import (
	"flag"
	"log"
	"net"

	"github.com/clevertechru/server_pow/internal/server"
	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/quotes"
)

func main() {
	configPath := flag.String("config", "config/server.yml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadServerConfig(*configPath)
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		cfg = config.DefaultServerConfig()
	}

	// Initialize quotes storage
	if err := quotes.Init(cfg.Server.Quotes.File); err != nil {
		log.Fatalf("Failed to initialize quotes: %v", err)
	}

	handler, err := server.NewHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	listener, err := net.Listen("tcp", cfg.Server.Host+":"+cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handler.ProcessConnection(conn)
	}
}
