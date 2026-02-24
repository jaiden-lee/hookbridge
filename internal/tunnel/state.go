package tunnel

import (
	"hookbridge/gen/tunnelv1"
	"log"
	"os"
	"sync"
	"time"
)

type tunnelStateStruct struct {
	clientsConnected           map[string]chan *tunnelv1.HttpRequest
	clientsConnectedLock       sync.RWMutex // no need to manually allocate
	mainServerResponseChan     chan *tunnelv1.HttpResponse
	clientsConnectedBufferSize int
	shutdownFunc               func()
	shutdownTimer              int
}

var tunnelState = tunnelStateStruct{
	clientsConnected:           map[string]chan *tunnelv1.HttpRequest{},
	mainServerResponseChan:     make(chan *tunnelv1.HttpResponse),
	clientsConnectedBufferSize: 32,  // probably won't ever get here,
	shutdownTimer:              120, // seconds
}

func (t *tunnelStateStruct) GetTunnelName() string {
	tunnelName, exists := os.LookupEnv("TUNNEL_NAME")
	if !exists {
		// then this container is broken, panic and exit out of container
		log.Fatal("No TUNNEL_NAME environment variable was provided")
	}

	return tunnelName
}

func SetShutdownFunc(fn func()) {
	tunnelState.shutdownFunc = fn
}

func DisconnectCountdown() {
	time.AfterFunc(time.Duration(tunnelState.shutdownTimer)*time.Second, func() {
		log.Println("Shutting down, since no clients have been connected in 2 minutes.")
		tunnelState.clientsConnectedLock.RLock()
		defer tunnelState.clientsConnectedLock.RUnlock()
		if len(tunnelState.clientsConnected) == 0 {
			tunnelState.shutdownFunc()
		}

	})
}

// TODO: add checks within main server in a goroutine to call docker wait to check if any error occurs??
// or perhaps add a sleep like 500ms to wait if it exits, and if not, then we're good to move on
