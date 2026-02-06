package main

import (
	"hookbridge/gen/tunnelv1"
	"log"

	"context"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatal("Failed to dial grpc server")
	}

	defer conn.Close()

	log.Println("connected to grpc server!")

	client := tunnelv1.NewTunnelServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = client.OpenTunnel(ctx) // temporary, not testing the stream yet

	if err != nil {
		log.Fatalf("failed to open tunnel/stream\n %s", err)
	}

	log.Println("tunnel/stream created and connected")
}
