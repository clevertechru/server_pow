package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clevertechru/server_pow/internal/server/service"
	"github.com/clevertechru/server_pow/pkg/config"
)

type HTTPServer struct {
	httpServer *http.Server
	handler    *service.RequestHandler
}

func NewHTTPServer(cfg *config.ServerConfig) (*HTTPServer, error) {
	handler, err := service.NewRequestHandler(cfg)
	if err != nil {
		return nil, err
	}

	server := &HTTPServer{
		handler: handler,
		httpServer: &http.Server{
			Addr:         net.JoinHostPort(cfg.Server.Host, cfg.Server.Port),
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}

	return server, nil
}

func (s *HTTPServer) Start() error {
	// Handle graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	log.Printf("Server started on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *HTTPServer) StartTLS(certFile, keyFile string) error {
	// Handle graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	log.Printf("Server started on %s (TLS)", s.httpServer.Addr)
	return s.httpServer.ListenAndServeTLS(certFile, keyFile)
}
