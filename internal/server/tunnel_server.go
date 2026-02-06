package server

import (
	"hookbridge/gen/tunnelv1"
	"log"

	"google.golang.org/grpc/metadata"
)

type MainServerTunnelStruct struct {
	tunnelv1.UnimplementedTunnelServiceServer
}

func (server *MainServerTunnelStruct) OpenTunnel(stream tunnelv1.TunnelService_OpenTunnelServer) error {
	var tunnelName string
	md, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		tunnelNames := md.Get("tunnel-name")
		if len(tunnelNames) > 0 {
			tunnelName = tunnelNames[0]
		} else {
			log.Printf("No tunnel name provided in metadata")
			return nil
		}
	}

	log.Printf("New tunnel connection from tunnel name: %s", tunnelName)
	// need a way to determine which client this is from?
	// we can't use random id
	// API needs to know which channel to forward this to
	// OR, I can do requests/response, and instead, we dial the other container?
	// and the other container runs a server?
	// or else i need like a welcome message?
	//
	return nil

}
