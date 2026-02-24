package main

import (
	"context"
	"hookbridge/gen/tunnelv1"
	"hookbridge/internal/tunnel"
	"log"
	"os"
	"os/signal"
	"syscall"

	"net"

	grpc "google.golang.org/grpc"
)

func main() {
	// -- Start GRPC client (connects to main server)
	// primary owner is sent as part of the forwarded request as a tag
	// when client receives, if it has that as primary owner, it will forward automatically
	mainCtx, mainCancel := context.WithCancel(context.Background())
	tunnel.SetShutdownFunc(mainCancel)
	tunnel.DisconnectCountdown()

	tunnelClientErr := make(chan error, 1)
	go func() {
		tunnelClientErr <- tunnel.InitTunnelClient(mainCtx)
	}()

	// -- Start GRPC Server (client streams)
	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	tunnelv1.RegisterTunnelServiceServer(grpcServer, &tunnel.TunnelServiceServerStruct{})

	log.Println("grpc server listening on port 50051")

	grpcServerErr := make(chan error, 1)
	go func() {
		err := grpcServer.Serve(lis)
		grpcServerErr <- err
	}()

	// --Listen for CTRL+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// --Listen for 1 of these to terminate first
	select {
	case err := <-tunnelClientErr:
		if err != nil {
			log.Printf("tunnel client exited: %v", err)
		} else {
			log.Printf("tunnel client exited cleanly")
		}

	case err := <-grpcServerErr:
		log.Printf("grpc server exited: %v", err)

	case sig := <-sigCh:
		log.Printf("received signal %v, shutting down", sig)

	case <-mainCtx.Done():
		log.Printf("Main context was cancelled; no clients connected")
	}

	mainCancel()              // shutdown the stream to mainserver if still running
	grpcServer.GracefulStop() // tunnel clients will auto shutdown
}
