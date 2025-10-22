package main

import (
	"log"

	"github.com/jaiden-lee/hookbridge/internal/server/routes"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("cmd/server/.env")
	if err != nil {
		log.Fatalf("Failed to load .env file")
	}
	router := routes.NewRouter()

	log.Println("HookBridge server running on :8080")

	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
