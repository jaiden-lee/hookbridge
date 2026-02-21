package tunnel

import (
	"fmt"
	"log"

	"context"
	"hookbridge/gen/tunnelv1"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// once this finishes, it should terminate client
func InitTunnelClient(mainCtx context.Context) error {
	// address temporary right now
	conn, err := grpc.NewClient(
		"host.docker.internal:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return fmt.Errorf("Failed to dial main server")
	}

	defer conn.Close()

	log.Println("connected to main server!")

	client := tunnelv1.NewTunnelServiceClient(conn)
	ctx, cancel := context.WithCancel(mainCtx)
	defer cancel()

	md := metadata.Pairs("tunnel-name", tunnelState.GetTunnelName())
	ctx = metadata.NewOutgoingContext(ctx, md) // add metadata to current context

	stream, err := client.OpenTunnel(ctx)
	if err != nil {
		return fmt.Errorf("failed to open tunnel/stream\n %s", err)
	}

	log.Println("tunnel to main server created and connected")

	//---------------------------------------------------------

	streamCtx, streamCancel := context.WithCancel(stream.Context())
	defer streamCancel()

	go ListenForHttpRequestFromMainServer(stream, streamCtx, streamCancel, mainCtx)

	// forwards response to main server
	for {
		select {
		case <-streamCtx.Done():
			return nil
		case <-mainCtx.Done():
			return nil
		case httpResponse, ok := <-tunnelState.mainServerResponseChan:
			if !ok {
				log.Printf("Channel closed, closing stream")
				return nil
			}

			err = stream.Send(httpResponse) // can just forward same one; response_id doesn't matter
			if err != nil {
				// main server closed stream, exiting
				log.Printf("Main server rejected HttpResponse, stream closed")
				return nil
			}
		}

	}
}

func ListenForHttpRequestFromMainServer(stream tunnelv1.TunnelService_OpenTunnelClient, streamCtx context.Context, streamCancel context.CancelFunc, mainCtx context.Context) {
	defer streamCancel()

	for {
		httpRequest, err := stream.Recv()
		// means stream closed, exit out of tunnel
		if err != nil {
			log.Printf("Main server closed stream")
			return
		}

		select {
		case <-streamCtx.Done():
			return
		case <-mainCtx.Done():
			return
		default:
		}

		// choose a random connected client as the PRIMARY;
		// the rest of the channels we need to send to as well (separate goroutine?)
		// to prevent HOL blocking, I should just use a long buffer; using 32 for now
		primaryClientID, success := GetRandomClientID()
		if !success {
			log.Printf("no clients connected, shutting down tunnel container")
			return
		}

		BroadcastHttpRequestToClientStreams(httpRequest, primaryClientID)
	}
}

func BroadcastHttpRequestToClientStreams(httpRequest *tunnelv1.HttpRequest, primaryClientID string) {
	tunnelState.clientsConnectedLock.RLock()
	defer tunnelState.clientsConnectedLock.RUnlock()

	for clientId, ch := range tunnelState.clientsConnected {
		var msg *tunnelv1.HttpRequest

		if clientId == primaryClientID {
			msg = proto.Clone(httpRequest).(*tunnelv1.HttpRequest)
			msg.AssignedToRespond = true
		} else {
			msg = httpRequest
		}

		// non-blocking send
		select {
		case ch <- msg:
		default:
			log.Printf("client %s queue full, dropping request", clientId)
		}
	}
}

func GetRandomClientID() (string, bool) {
	tunnelState.clientsConnectedLock.RLock()
	defer tunnelState.clientsConnectedLock.RUnlock()

	for k := range tunnelState.clientsConnected {
		return k, true // first one encountered
	}
	return "", false // map was empty
}
