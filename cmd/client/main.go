package main

import (
	"log"
	"time"

	"github.com/clevertechru/server_pow/internal/client"
	"github.com/clevertechru/server_pow/pkg/config"
)

func main() {
	config := config.NewClientConfig()
	handler := client.NewHandler(config)

	for {
		if err := handler.MakeRequest(); err != nil {
			log.Printf("Error: %v", err)
		}
		time.Sleep(config.RequestsDelay)
	}
}
