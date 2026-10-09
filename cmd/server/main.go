package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/thakur-vijay/geotrack-backend/internal/app"
	"github.com/thakur-vijay/geotrack-backend/internal/config"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()

	if err := application.Run(cfg.Port); err != nil {
		log.Fatal(err)
	}
}
