package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaiden-lee/hookbridge/internal/server/utils"
)

type authHandlersStruct struct{}

var AuthHandlers = &authHandlersStruct{}

func (s *authHandlersStruct) LogoutHandler(c *gin.Context) {
	user, err := getUserFromContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	go utils.AuthService.SignOutUser(user.SessionID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Signed out successfully!",
	})
}
