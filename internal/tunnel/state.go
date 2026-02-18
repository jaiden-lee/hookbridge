package tunnel

import (
	"hookbridge/gen/tunnelv1"
	"log"
	"os"
	"sync"
)

type tunnelStateStruct struct {
	clientsConnected           map[string]chan *tunnelv1.HttpRequest
	clientsConnectedLock       sync.RWMutex // no need to manually allocate
	mainServerResponseChan     chan *tunnelv1.HttpResponse
	clientsConnectedBufferSize int
}

var tunnelState = tunnelStateStruct{
	clientsConnected:           map[string]chan *tunnelv1.HttpRequest{},
	mainServerResponseChan:     make(chan *tunnelv1.HttpResponse),
	clientsConnectedBufferSize: 32, // probably won't ever get here
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
