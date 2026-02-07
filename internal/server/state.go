package server

type serverStateStruct struct {
	activeTunnels map[string]int // tunnelName -> port number
	serverIp      string
}

var serverState = serverStateStruct{
	activeTunnels: map[string]int{},
	serverIp:      "localhost",
}
