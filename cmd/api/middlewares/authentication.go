package middlewares

import (
	"strings"

	"github.com/Farmaan-Malik/gollama-app/utils"
	"github.com/gin-gonic/gin"
)

func Authentication(ctx *gin.Context) {
	token := ctx.Request.Header.Get("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(400, gin.H{"success": false, "message": "missing auth header"})
	}
	prefix := "Bearer "
	ok := strings.HasPrefix(token, prefix)
	if !ok {
		ctx.AbortWithStatusJSON(400, gin.H{"success": false, "message": "invalid token format"})
		return
	}
	trimmedToken := strings.TrimPrefix(token, prefix)
	err := utils.ValidateJwt(trimmedToken)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"success": false, "message": "hereee"})
		return
	}
	ctx.Next()
}
