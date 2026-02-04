package tunnel

import (
	"hookbridge/gen/tunnelv1"
	"log"
	"os"
)

type tunnelStateStruct struct {
	clientsConnected       map[string]chan *tunnelv1.HttpResponse
	mainServerResponseChan chan *tunnelv1.HttpResponse
}

var tunnelState = tunnelStateStruct{
	clientsConnected:       map[string]chan *tunnelv1.HttpResponse{},
	mainServerResponseChan: make(chan *tunnelv1.HttpResponse),
}

func (t *tunnelStateStruct) GetTunnelName() string {
	tunnelName, exists := os.LookupEnv("TUNNEL_NAME")
	if !exists {
		// then this container is broken, panic and exit out of container
		log.Fatal("No TUNNEL_NAME environment variable was provided")
	}

	return tunnelName
}

// TODO: add checks within main server in a goroutine to call docker wait to check if any error occurs??
// or perhaps add a sleep like 500ms to wait if it exits, and if not, then we're good to move on
