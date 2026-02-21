package server

import (
	"encoding/json"
	"hookbridge/internal/api"
	"strings"

	"log"
	"regexp"

	"hookbridge/gen/tunnelv1"

	"github.com/gin-gonic/gin"
)

type connectHandlersStruct struct{}
type tunnelHandlersStruct struct{}

var connectHandlers = connectHandlersStruct{}

func (_ *connectHandlersStruct) connectOrCreateTunnel(c *gin.Context) {
	var requestBody api.ConnectToTunnelRequest

	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil {
		log.Println("json request body was invalid")
		log.Println(err)
		c.AbortWithStatusJSON(400, gin.H{
			"error": "invalid json request body",
		})
		return
	}

	if requestBody.TunnelName == "" {
		log.Println("tunnel_name parameter was empty")
		c.AbortWithStatusJSON(400, gin.H{
			"error": "tunnel_name parameter is empty",
		})
		return
	}

	if !isValidTunnelName(requestBody.TunnelName) {
		log.Println("tunnel name was invalid; no special characters allowed")
		c.AbortWithStatusJSON(400, gin.H{
			"error": "invalid tunnel name, no spaces or special characters besides _ and - are allowed",
		})
		return
	}

	tunnelName := strings.ToLower(requestBody.TunnelName)

	log.Printf("Connecting to tunnel <%s>\n", tunnelName)

	port, tunnelExists := serverState.activeTunnels[tunnelName]

	if !tunnelExists {
		log.Println("Tunnel doesn't exist. Creating tunnel...")

		// this is created beforehand, since tunnel dials server immediately; so needs to exist in map beforehand
		serverState.tunnelRequestChannels[tunnelName] = make(chan *tunnelv1.HttpRequest)
		serverState.tunnelResponseCHannels[tunnelName] = make(chan *tunnelv1.HttpResponse)
		port, err = startTunnel(tunnelName)
		serverState.activeTunnels[tunnelName] = port

		if err != nil {
			log.Println("Failed to start tunnel docker container")
			log.Println(err)

			cleanupTunnel(tunnelName)

			c.AbortWithStatusJSON(500, gin.H{
				"error": "internal server error, failed to start tunnel docker container",
			})
			return
		}

	}

	log.Printf("Tunnel is ready. Returning port number %d to client...\n", port)

	c.JSON(200, api.ConnectToTunnelResponse{
		TunnelIp: serverState.serverIp,
		Port:     port,
	})
}

func isValidTunnelName(tunnelName string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`) // _, -, abc..., 123...    no space or special chars
	return re.MatchString(tunnelName)
}

func cleanupTunnel(tunnelName string) {
	_, ok := serverState.tunnelRequestChannels[tunnelName]
	if ok {
		// close(requestChan) let be garbage collected
		delete(serverState.tunnelRequestChannels, tunnelName)
	}

	_, ok = serverState.tunnelResponseCHannels[tunnelName]
	if ok {
		// close(responseChan)
		// don't close responseChan; let it be garbage collected
		delete(serverState.tunnelResponseCHannels, tunnelName)
	}

	if _, ok = serverState.activeTunnels[tunnelName]; ok {
		delete(serverState.activeTunnels, tunnelName) // port number
	}
}
