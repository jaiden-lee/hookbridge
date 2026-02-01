package server

import (
	"encoding/json"
	"hookbridge/internal/api"

	"log"

	"github.com/gin-gonic/gin"
)

type connectHandlersStruct struct{}
type tunnelHandlersStruct struct{}

var connectHandlers = connectHandlersStruct{}

func (_ *connectHandlersStruct) connectOrCreateTunnel(c *gin.Context) {
	var requestBody api.ConnectToTunnelRequest

	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil || requestBody.TunnelName == "" { // body doesn't match
		c.AbortWithStatusJSON(400, gin.H{
			"error": "invalid json request body",
		})
		return
	}

	log.Printf("Connecting to tunnel <%s>\n", requestBody.TunnelName)

	_, tunnelExists := serverState.activeTunnels[requestBody.TunnelName]

	if !tunnelExists {
		log.Println("Tunnel doesn't exist. Creating tunnel...")
		serverState.activeTunnels[requestBody.TunnelName] = true
		err := startTunnel()
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{
				"error": "internal server error, failed to start tunnel docker container",
			})
			return
		}
	}

}
