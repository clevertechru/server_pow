package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/clevertechru/server_pow/internal/server"
	"github.com/clevertechru/server_pow/pkg/config"
)

func main() {
	cfg := config.NewServerSettings()
	handler := server.NewHandler(cfg)

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
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
