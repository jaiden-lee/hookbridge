package handlers

import (
	"github.com/gin-gonic/gin"
)

type authHandlersStruct struct{}

var AuthHandlers = &authHandlersStruct{}

func (s *authHandlersStruct) LogoutHandler(c *gin.Context) {

}
