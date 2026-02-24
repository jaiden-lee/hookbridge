package server

import (
	"hookbridge/gen/tunnelv1"
	"log"

	"context"

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

	// need to use a channel for the tunnel name
	// when an API request is received on the wildcard endpoint, send signal to channel
	// causes this thread to wake up and then forward that message through stream
	requestChannel, requestExists := serverState.tunnelRequestChannels[tunnelName]
	// responseChannel, responseExists := serverState.tunnelResponseCHannels[tunnelName]
	defer close(requestChannel)
	defer cleanupTunnel(tunnelName)

	if !requestExists {
		return nil // channel doesn't exist, should never happen since API should create channel before client can connect to it
	}

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go tunnelResponseReceiver(stream, tunnelName, ctx, cancel)
	for {
		select {
		case <-ctx.Done():
			return nil

		case httpRequest, ok := <-requestChannel:
			if !ok {
				log.Printf("Request channel for tunnel %s closed", tunnelName)
				return nil
			}
			err := stream.Send(httpRequest)
			if err != nil {
				log.Printf("Error sending message through tunnel stream for tunnel %s: %v", tunnelName, err)
				return err
			}
		}
	}
}

func tunnelResponseReceiver(stream tunnelv1.TunnelService_OpenTunnelServer, tunnelName string, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	// defer close(responseChannel)

	for {
		httpResponse, err := stream.Recv()
		if err != nil {
			log.Printf("Tunnel %s disconnected", tunnelName)
			return
		}

		requestId := httpResponse.RequestId
		responseChannel, responseExists := serverState.tunnelResponseCHannels[requestId]
		if !responseExists {
			log.Printf("No response channel for requestId, skipping request")
			continue
		}

		select { // make cancellation deterministic, so it runs first; prevents race condition if both are ready at same time
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			// close from this thread, since it is the only goroutine that sends
			return
		case responseChannel <- httpResponse: // nonblocking operation, should send immediately, unless ctx.Done is ready
		}

	}
}
