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

	tunnelState.clientsConnected[tunnelId] = make(chan int)

	ListenForClientMessage(stream)

	return nil
}

func ListenForClientMessage(stream tunnelv1.TunnelService_OpenTunnelServer) error {
	for {
		message, err := stream.Recv()
		if err != nil {
			if isNormalDisconnect(err) {
				return nil
			}
			return err
		}

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
