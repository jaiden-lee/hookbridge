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
	return nil
}
