package main

import (
	"hookbridge/gen/tunnelv1"
	"hookbridge/internal/tunnel"

	"net"

	grpc "google.golang.org/grpc"
)

func main() {
	// primary owner is sent as part of the forwarded request as a tag
	// when client receives, if it has that as primary owner, it will forward automatically
	lis, _ := net.Listen("tcp", "50051")

	grpcServer := grpc.NewServer()
	tunnelv1.RegisterTunnelServiceServer(grpcServer, &tunnel.TunnelServiceServerStruct{})

	grpcServer.Serve(lis)
}
