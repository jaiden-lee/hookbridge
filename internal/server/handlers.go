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

	_, exists := serverState.activeTunnels[requestBody.TunnelName]

	if !exists {
		// create tunnel
		log.Println("Tunnel doesn't exist. Creating tunnel...")

	}
}
