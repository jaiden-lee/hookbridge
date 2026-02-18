package tunnel

import (
	"log"

	"context"
	"hookbridge/gen/tunnelv1"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func InitTunnelClient() {
	// address temporary right now
	conn, err := grpc.NewClient(
		"host.docker.internal:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatal("Failed to dial main server")
	}

	defer conn.Close()

	log.Println("connected to main server!")

	client := tunnelv1.NewTunnelServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	md := metadata.Pairs("tunnel-name", tunnelState.GetTunnelName())
	ctx = metadata.NewOutgoingContext(ctx, md) // add metadata to current context

	stream, err := client.OpenTunnel(ctx)
	if err != nil {
		log.Fatalf("failed to open tunnel/stream\n %s", err)
	}

	log.Println("tunnel to main server created and connected")

	//---------------------------------------------------------

	streamCtx, streamCancel := context.WithCancel(stream.Context())
	defer streamCancel()

	go ListenForHttpRequestFromMainServer(stream, streamCtx, streamCancel)

	// forwards response to main server
	for {
		select {
		case <-streamCtx.Done():
			return
		case httpResponse, ok := <-tunnelState.mainServerResponseChan:
			if !ok {
				log.Printf("Channel closed, closing stream")
				return
			}

			err = stream.Send(httpResponse) // can just forward same one; response_id doesn't matter
			if err != nil {
				// main server closed stream, exiting
				log.Printf("Main server rejected HttpResponse, stream closed")
				return
			}
		}

	}
}

func ListenForHttpRequestFromMainServer(stream tunnelv1.TunnelService_OpenTunnelClient, streamCtx context.Context, streamCancel context.CancelFunc) {
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
		default:
		}

		// choose a random connected client as the PRIMARY;
		// the rest of the channels we need to send to as well (separate goroutine?)
		// to prevent HOL blocking, I should just use a long buffer; using 32 for now
		primaryClientID, success := GetRandomClientID()
		if !success {
			log.Printf("no clients connected, closing stream")
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
