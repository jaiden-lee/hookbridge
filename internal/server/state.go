package server

type serverStateStruct struct {
	activeTunnels map[string]bool
}

var serverState = serverStateStruct{
	activeTunnels: map[string]bool{},
}
