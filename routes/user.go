package routes

import (
	"example/gollama-app/store"
	"fmt"

	"github.com/gin-gonic/gin"
)

func SignupUserHandler(ctx *gin.Context) {
	var u store.User
	err := ctx.ShouldBindJSON(&u)
	if err != nil {
		fmt.Println("error while binding json", err)
		ctx.JSON(400, gin.H{"success": false, "message": err})
		return
	}
	doc, err := u.CreateUser()
	if err != nil {
		fmt.Println("error while creating document ")
		ctx.JSON(400, gin.H{"success": false, "message": fmt.Sprint(err)})
		return
	}
	ctx.JSON(201, gin.H{"success": true, "data": doc})
}

func LoginUserHandler(ctx *gin.Context) {
	var payload store.LoginPayload
	err := ctx.ShouldBindJSON(&payload)
	if err != nil {
		ctx.JSON(401, gin.H{"success": false, "message": fmt.Sprint(err)})
		return
	}
	_, err = payload.LoginUser()
	if err != nil {
		ctx.JSON(401, gin.H{"success": false, "message": fmt.Sprint(err)})
		return
	}
	ctx.JSON(200, gin.H{"success": true, "message": "user logged in successfully"})

}
