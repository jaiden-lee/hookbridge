package tunnel

import (
	"hookbridge/gen/tunnelv1"
	"log"
)

type TunnelServiceServerStruct struct {
	tunnelv1.UnimplementedTunnelServiceServer
}

// alias TunnelService_OpenTunnelServer = grpc.BidiStreamingServer[HttpResponse, HttpRequest]
func (server *TunnelServiceServerStruct) OpenTunnel(stream tunnelv1.TunnelService_OpenTunnelServer) error {
	log.Println("Tunnel opened with client")
	return nil
}
