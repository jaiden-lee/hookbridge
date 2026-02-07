package api

type ConnectToTunnelRequest struct {
	TunnelName string `json:"tunnel_name"`
}

type ConnectToTunnelResponse struct {
	TunnelIp string `json:"tunnel_ip"`
	Port     int    `json:"port"`
}
