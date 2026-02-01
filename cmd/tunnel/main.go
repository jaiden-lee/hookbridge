package main

import (
	"hookbridge/gen/tunnelv1"
	"hookbridge/internal/tunnel"
	"log"

	"net"

	grpc "google.golang.org/grpc"
)

func main() {
	// primary owner is sent as part of the forwarded request as a tag
	// when client receives, if it has that as primary owner, it will forward automatically
	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	tunnelv1.RegisterTunnelServiceServer(grpcServer, &tunnel.TunnelServiceServerStruct{})

	log.Println("grpc server listening on port 50051")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
