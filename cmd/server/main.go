package main

import (
	"log"

	"github.com/jaiden-lee/hookbridge/internal/server/routes"
)

func main() {
	router := routes.NewRouter()

	log.Println("HookBridge server running on :8080")

	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
