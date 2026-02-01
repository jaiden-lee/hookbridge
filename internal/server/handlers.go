package server

import (
	"encoding/json"
	"hookbridge/internal/api"

	"github.com/gin-gonic/gin"
)

type connectHandlersStruct struct{}
type tunnelHandlersStruct struct{}

func (_ *connectHandlersStruct) connectOrCreateTunnel(c *gin.Context) {
	var requestBody api.ConnectToTunnelRequest

	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil || requestBody.TunnelName == "" { // body doesn't match
		c.AbortWithStatusJSON(400, gin.H{
			"error": "invalid json request body",
		})
	}
}
