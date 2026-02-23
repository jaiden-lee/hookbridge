package server

import (
	"hookbridge/gen/tunnelv1"
)

type serverStateStruct struct {
	activeTunnels          map[string]int // tunnelName -> port number
	tunnelRequestChannels  map[string]chan *tunnelv1.HttpRequest
	tunnelResponseCHannels map[string]chan *tunnelv1.HttpResponse
	serverIp               string
	requestTimeout         int // seconds
}

var serverState = serverStateStruct{
	activeTunnels:          map[string]int{},
	tunnelRequestChannels:  map[string]chan *tunnelv1.HttpRequest{},
	tunnelResponseCHannels: map[string]chan *tunnelv1.HttpResponse{},
	serverIp:               "localhost",
	requestTimeout:         300,
}
