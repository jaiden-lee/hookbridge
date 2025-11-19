package grpc

import (
	"sync"

	"github.com/jaiden-lee/hookbridge/proto"
)

type PendingRequests struct {
	mu               sync.Mutex
	flowIDChannelMap map[string](chan *proto.HTTPResponse)
	// map that takes in string as input, and maps to a channel that expects pointer to HTTPResposne
}

// this should not be its own goroutine. It should use the goroutine of the STREAM
// this is so that we can return the error message to the higher level STREAM function
func StartServerReceiver(stream proto.BridgeService_StreamServer, pendingRequests *PendingRequests) error {
	for {
		message, err := stream.Recv()
		if err != nil {
			return err
		}

		// message is an HTTP response. Send to channel
		pendingRequests.mu.Lock()
		channel := pendingRequests.flowIDChannelMap[message.FlowId]
		if channel != nil {
			delete(pendingRequests.flowIDChannelMap, message.FlowId) // delete from map asap
		}
		pendingRequests.mu.Unlock()

		// in case channel doesn't exist in map
		if channel != nil {
			channel <- message
			// Use a buffered channel of size 1 (instead of unbuffered channel)
			// This prevents this thread from blocking; it only blocks if buffer is full
			// BUT, we will never send again, so buffer will never be full (if we try to send)

			// close the channel and delete from map now to prevent memory leak
			// Oh but what if other side hasn't read from channel?
			// If the other side was already waiting to read, send goes through immediately
			// So it is safe to close the channel now.Otherwise, if the other side hasn't started listening yet
			// Then this means it probably never will (since it should start listening asap)
			close(channel)
		}
	}
}
