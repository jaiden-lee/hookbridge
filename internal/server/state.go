package server

type serverStateStruct struct {
	activeTunnels map[string]int // tunnelName -> port number
}

var serverState = serverStateStruct{
	activeTunnels: map[string]int{},
}
