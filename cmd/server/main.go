package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

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

	// Initialize quotes storage)
	if err := quotes.Init(cfg.Server.Quotes.File); err != nil {
		log.Fatalf("Failed to initialize quotes storage: %v", err)
	}

	handler, err := server.NewHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	addr := net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		cancel()

		if err := listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}

		// Shutdown worker pool
		handler.Shutdown()
	}()

	log.Printf("Server started on %s", addr)

	// Accept connections until context is canceled
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
			handler.ProcessConnection(conn)
		}
	}
}
