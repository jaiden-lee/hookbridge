package tunnel

import (
	"log"

	"context"
	"hookbridge/gen/tunnelv1"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitTunnelClient() {
	// address temporary right now
	conn, err := grpc.NewClient(
		"host.docker.internal:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatal("Failed to dial main server")
	}

	defer conn.Close()

	log.Println("connected to main server!")

	client := tunnelv1.NewTunnelServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.OpenTunnel(ctx)
	if err != nil {
		log.Fatalf("failed to open tunnel/stream\n %s", err)
	}

	log.Println("tunnel to main server created and connected")

	for {
		httpResponse := <-tunnelState.mainServerResponseChan
		stream.Send(httpResponse) // can just forward same one; response_id doesn't matter
	}
}
