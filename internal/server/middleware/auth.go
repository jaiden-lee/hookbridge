package middleware

import (
	"net/http"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jaiden-lee/hookbridge/internal/server/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		// return empty string if doesn't exist
		if authHeader == "" {
			// 401 error
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "missing Authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "missing Bearer field",
			})
			return
		}

		token := parts[1]
		userData, err := utils.AuthService.VerifyJWT(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired access token",
			})
		}

		ctx.Set("user", userData)

		ctx.Next()
	}
}

func GetUserFromCtx(c *gin.Context) (*utils.UserData, bool) {
	val, ok := c.Get("user")
	if !ok || val == nil {
		return nil, false
	}
	user, ok := val.(*utils.UserData)
	if !ok {
		return nil, false
	}
	return user, true
}
