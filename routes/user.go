package routes

import (
	"context"
	"fmt"

	"github.com/Farmaan-Malik/gollama-app/store"

	"github.com/gin-gonic/gin"
)

type Api struct {
	Store *store.Store
}

func (a *Api) SignupUserHandler(ctx *gin.Context) {
	var u *store.User
	err := ctx.ShouldBindJSON(&u)
	if err != nil {
		fmt.Println("error while binding json", err)
		ctx.JSON(400, gin.H{"success": false, "message": err})
		return
	}
	doc, err := a.Store.CreateUser(u)
	if err != nil {
		fmt.Println("error while creating document ")
		ctx.JSON(400, gin.H{"success": false, "message": fmt.Sprint(err), "data": doc})
		return
	}
	ctx.JSON(201, gin.H{"success": true, "data": doc})
}

func (a *Api) LoginUserHandler(ctx *gin.Context) {
	var payload *store.LoginPayload
	err := ctx.ShouldBindJSON(&payload)
	if err != nil {
		ctx.JSON(401, gin.H{"success": false, "message": fmt.Sprint(err)})
		return
	}
	_, err = a.Store.LoginUser(payload)
	if err != nil {
		ctx.JSON(401, gin.H{"success": false, "message": fmt.Sprint(err)})
		return
	}
	ctx.JSON(200, gin.H{"success": true, "message": "user logged in successfully"})

}

func (a *Api) GetInitialDataHandler(ctx *gin.Context) {
	var payload *store.InititalPrompt
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(401, gin.H{"success": false, "message": "incorrect data format"})
		return
	}
	if err := a.Store.GetInitialData(payload); err != nil {
		ctx.JSON(500, gin.H{"success": false, "message": "error saving initial data"})
		return
	}
	data, err := a.Store.Redis.HGetAll(ctx, payload.UserId).Result()
	if err != nil {
		ctx.JSON(500, gin.H{"success": false, "message": "error fetching data from redis"})
		return
	}
	fmt.Println("Data: ", data)
	ctx.JSON(200, gin.H{"success": true, "message": "initial data recieved", "data": data})
}

func (a *Api) GetQuestionHandler(ctx *gin.Context) {
	var payload *store.Ask
	err := ctx.ShouldBindJSON(&payload)
	if err != nil {
		ctx.JSON(401, gin.H{"success": false, "message": "incorrect data format"})
		return
	}
	question, err := a.Store.GetQuestion(context.Background(), payload)
	if err != nil {
		ctx.JSON(500, gin.H{"success": false, "message": err})
		return
	}
	ctx.JSON(200, gin.H{"success": true, "data": question})
}
