package tunnel

import (
	"hookbridge/gen/tunnelv1"
	"log"

	"context"
	"errors"
	"io"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TunnelServiceServerStruct struct {
	tunnelv1.UnimplementedTunnelServiceServer
}

// alias TunnelService_OpenTunnelServer = grpc.BidiStreamingServer[HttpResponse, HttpRequest]
func (server *TunnelServiceServerStruct) OpenTunnel(stream tunnelv1.TunnelService_OpenTunnelServer) error {
	log.Println("Tunnel opened with client")

	tunnelIdUUID, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	tunnelId := tunnelIdUUID.String()

	responseChannel := make(chan *tunnelv1.HttpResponse)
	defer close(responseChannel)

	tunnelState.clientsConnected[tunnelId] = responseChannel

	ListenForClientMessage(stream) // blocking

	return nil
}

// goroutine that sends channel is same that closes channel; no need to check if closed
func ListenForClientMessage(stream tunnelv1.TunnelService_OpenTunnelServer) error {
	for {
		messageHttpResponse, err := stream.Recv()
		if err != nil {
			if isNormalDisconnect(err) {
				return nil
			}
			return err
		}

		// no need to care about if we are the main client, since only one client ever sends
		tunnelState.mainServerResponseChan <- messageHttpResponse
	}
}

func isNormalDisconnect(err error) bool {
	if err == nil {
		return false
	}

	// Client cleanly closed its send side
	if errors.Is(err, io.EOF) {
		return true
	}

	// Context cancellation / timeout
	if errors.Is(err, context.Canceled) {
		return true
	}

	// gRPC status code-based disconnects
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Canceled:
		// Client canceled (Ctrl+C, Close, etc.)
		return true

	case codes.Unavailable:
		// Transport closed (client disconnected, network drop)
		return true

	case codes.DeadlineExceeded:
		// Only treat as normal if YOU expected timeout behavior
		// Usually not normal for tunnels
		return false
	}

	return false
}
