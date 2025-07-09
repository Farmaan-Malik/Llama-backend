package middlewares

import (
	"fmt"
	"strings"

	"github.com/Farmaan-Malik/gollama-app/utils"
	"github.com/gin-gonic/gin"
)

func Authentication(ctx *gin.Context) {
	token := ctx.Request.Header.Get("Authorization")
	if token == "" {
		fmt.Println("Missing Auth Header")
		ctx.AbortWithStatusJSON(400, gin.H{"success": false, "message": "missing auth header"})
	}
	prefix := "Bearer "
	ok := strings.HasPrefix(token, prefix)
	if !ok {
		fmt.Println("Invalid Auth Header Format")
		ctx.AbortWithStatusJSON(400, gin.H{"success": false, "message": "invalid token format"})
		return
	}
	trimmedToken := strings.TrimPrefix(token, prefix)
	err := utils.ValidateJwt(trimmedToken)
	if err != nil {
		fmt.Println("Invalid JWT Token")
		ctx.AbortWithStatusJSON(400, gin.H{"success": false, "message": "invalid jwt token"})
		return
	}
	ctx.Next()
}
