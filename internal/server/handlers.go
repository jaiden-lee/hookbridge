package server

import (
	"encoding/json"
	"hookbridge/internal/api"

	"log"
	"regexp"

	"github.com/gin-gonic/gin"
)

type connectHandlersStruct struct{}
type tunnelHandlersStruct struct{}

var connectHandlers = connectHandlersStruct{}

func (_ *connectHandlersStruct) connectOrCreateTunnel(c *gin.Context) {
	var requestBody api.ConnectToTunnelRequest

	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil || requestBody.TunnelName == "" { // body doesn't match
		log.Println("json request body was invalid")
		c.AbortWithStatusJSON(400, gin.H{
			"error": "invalid json request body",
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

	log.Printf("Connecting to tunnel <%s>\n", requestBody.TunnelName)

	_, tunnelExists := serverState.activeTunnels[requestBody.TunnelName]

	if !tunnelExists {
		log.Println("Tunnel doesn't exist. Creating tunnel...")
		serverState.activeTunnels[requestBody.TunnelName] = true
		err := startTunnel(requestBody.TunnelName)
		if err != nil {
			log.Println("Failed to start tunnel docker container")
			log.Println(err)

			c.AbortWithStatusJSON(500, gin.H{
				"error": "internal server error, failed to start tunnel docker container",
			})
			return
		}
	}

}

func isValidTunnelName(tunnelName string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`) // _, -, abc..., 123...    no space or special chars
	return re.MatchString(tunnelName)
}
