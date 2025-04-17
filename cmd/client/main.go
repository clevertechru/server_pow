package main

import (
	"flag"
	"log"
	"time"

	"github.com/clevertechru/server_pow/internal/client"
	"github.com/clevertechru/server_pow/pkg/config"
)

func main() {
	configPath := flag.String("config", "config/client.yml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadClientConfig(*configPath)
	if err != nil {
		log.Printf("failed to read config file: %v", err)
		cfg = config.DefaultClientConfig()
	}
	handler := client.NewHandler(cfg)

	for {
		if err := handler.MakeRequest(); err != nil {
			log.Printf("Error: %v", err)
		}
		delay, err := time.ParseDuration(cfg.Client.RequestDelay)
		if err != nil {
			log.Printf("Error parsing delay: %v", err)
			delay = 100 * time.Millisecond
		}
		time.Sleep(delay)
	}
}
