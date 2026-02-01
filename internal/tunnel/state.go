package tunnel

type tunnelStateStruct struct {
	clientsConnected map[string]chan
}

var tunnelState = tunnelStateStruct{
	clientsConnected: map[string]chan{},
}
