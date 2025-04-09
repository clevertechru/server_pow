package main

import (
	"github.com/clevertechru/server_pow/internal/client"
	"github.com/clevertechru/server_pow/pkg/config"
	"log"
	"time"
)

func main() {
	cfg := config.ClientConfigNew()
	handler := client.NewHandler(cfg)

	for {
		if err := handler.MakeRequest(); err != nil {
			log.Printf("Error: %v", err)
		}
		time.Sleep(cfg.RequestsDelayMs)
	}
}
