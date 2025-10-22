package handlers

import (
	"github.com/gin-gonic/gin"
)

type connectHandlersStruct struct{}

var ConnectHandlers = &connectHandlersStruct{}

func (s *connectHandlersStruct) ConnectToProjectHandler(c *gin.Context) {

}

func (s *connectHandlersStruct) GetConnectionStatusHandler(c *gin.Context) {

}
