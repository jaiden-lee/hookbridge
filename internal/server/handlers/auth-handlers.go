package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaiden-lee/hookbridge/internal/server/utils"
	"github.com/jaiden-lee/hookbridge/pkg/api"
)

type authHandlersStruct struct{}

var AuthHandlers = &authHandlersStruct{}

// has middleware
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

// no auth middleware
func (s *authHandlersStruct) ExchangeRefreshTokenHandler(c *gin.Context) {
	var request api.ExchangeRefreshTokenRequest

	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	refreshResponse, err := utils.AuthService.ExchangeRefreshToken(request.RefreshToken)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, api.ExchangeRefreshTokenResponse{
		RefreshToken: refreshResponse.RefreshToken,
		AccessToken:  refreshResponse.AccessToken,
	})
}
