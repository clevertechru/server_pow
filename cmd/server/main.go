package main

import (
	"log"
	"net"

	"github.com/clevertechru/server_pow/internal/server"
	"github.com/clevertechru/server_pow/pkg/config"
)

func main() {
	cfg := config.ServerConfigNew()
	handler := server.NewHandler(cfg)

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}()

	log.Printf("Server started on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handler.HandleConnection(conn)
	}
}
