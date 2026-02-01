package tunnel

import "hookbridge/gen/tunnelv1"

type tunnelStateStruct struct {
	clientsConnected       map[string]chan *tunnelv1.HttpResponse
	mainServerResponseChan chan *tunnelv1.HttpResponse
}

var tunnelState = tunnelStateStruct{
	clientsConnected:       map[string]chan *tunnelv1.HttpResponse{},
	mainServerResponseChan: make(chan *tunnelv1.HttpResponse),
}
