package server

import (
	"encoding/json"
	"hookbridge/internal/api"
	"strings"

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

	_, tunnelExists := serverState.activeTunnels[tunnelName]

	if !tunnelExists {
		log.Println("Tunnel doesn't exist. Creating tunnel...")
		serverState.activeTunnels[tunnelName] = true
		err := startTunnel(tunnelName)

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
