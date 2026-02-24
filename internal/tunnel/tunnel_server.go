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

	requestChannel := make(chan *tunnelv1.HttpRequest, tunnelState.clientsConnectedBufferSize)
	tunnelState.clientsConnectedLock.Lock()
	tunnelState.clientsConnected[tunnelId] = requestChannel
	tunnelState.clientsConnectedLock.Unlock()

	defer disconnectClient(tunnelId)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go SendRequestToClientThread(stream, ctx, cancel, requestChannel)

	err = ListenForClientMessage(stream, ctx) // blocking
	return err
}

func SendRequestToClientThread(stream tunnelv1.TunnelService_OpenTunnelServer, ctx context.Context, cancel context.CancelFunc, requestChannel chan *tunnelv1.HttpRequest) {
	defer cancel()
	for {
		// listen for either incoming message, or stream is closed/context done
		select {
		case <-ctx.Done():
			log.Println("Stream closed, client is disconnected")
			return
		case httpRequest, ok := <-requestChannel:
			if !ok {
				log.Println("Request channel closed, means client was disconnected")
				return
			}
			// this httpRequest may be primary or not
			stream.Send(httpRequest)
		}
	}
}

// goroutine that sends channel is same that closes channel; no need to check if closed
func ListenForClientMessage(stream tunnelv1.TunnelService_OpenTunnelServer, ctx context.Context) error {
	for {
		messageHttpResponse, err := stream.Recv()
		if err != nil {
			if isNormalDisconnect(err) {
				return nil
			}
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case tunnelState.mainServerResponseChan <- messageHttpResponse: // unbuffered, will wait, so need to add to this select statement, since we're waiting on both
			// no need to care about if we are the main client, since only one client ever sends
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

func disconnectClient(tunnelId string) {
	tunnelState.clientsConnectedLock.Lock()
	defer tunnelState.clientsConnectedLock.Unlock()

	_, ok := tunnelState.clientsConnected[tunnelId]
	if ok {
		delete(tunnelState.clientsConnected, tunnelId) // remove first
	}

	if ok {
		// close(ch)
		// closing not necessary SINCE CHANNELS ARE GARBAGE COLLECTED
		// not closing will prevent race condition for tunnel_client.go when it broadcasts HttpRequests across all channels
		// well technically it would be fine becasue it's all locked behind a mutex
	}

	// also check if this is the last clientsConnected; if all clients connected are 0, then shutdown container too
	DisconnectCountdown()
}
