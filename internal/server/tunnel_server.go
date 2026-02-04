package server

import (
	"hookbridge/gen/tunnelv1"
	"log"
)

type MainServerTunnelStruct struct {
	tunnelv1.UnimplementedTunnelServiceServer
}

func (server *MainServerTunnelStruct) OpenTunnel(stream tunnelv1.TunnelService_OpenTunnelServer) error {
	log.Println("Main server tunnel opened")
	// need a way to determine which client this is from?
	// we can't use random id
	// API needs to know which channel to forward this to
	// OR, I can do requests/response, and instead, we dial the other container?
	// and the other container runs a server?
	// or else i need like a welcome message?
	// 
	return nil
	
}
