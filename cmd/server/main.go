package main

import (
	"hookbridge/gen/tunnelv1"
	"hookbridge/internal/server"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	go StartGRPCServer()

	router := server.GetRouter()
	router.Run()
}

// run in goroutine
func StartGRPCServer() {
	lis, err := net.Listen("tcp", ":8081")

	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	tunnelv1.RegisterTunnelServiceServer(grpcServer, &server.MainServerTunnelStruct{})

	log.Println("grpc server listening on port 50051")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
