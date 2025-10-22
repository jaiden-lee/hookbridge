package handlers

import (
	"github.com/gin-gonic/gin"
)

type tunnelHandlersStruct struct{}

var TunnelHandlers = &tunnelHandlersStruct{}

func (s *tunnelHandlersStruct) WebhookForwardingHandler(c *gin.Context) {

}
