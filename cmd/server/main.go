package main

import (
	"hookbridge/internal/server"
)

func main() {
	router := server.GetRouter()
	router.Run()
}
