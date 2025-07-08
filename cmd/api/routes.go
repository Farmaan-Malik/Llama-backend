package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Farmaan-Malik/gollama-app/cmd/api/middlewares"
	"github.com/Farmaan-Malik/gollama-app/internals/store"

	"github.com/gin-gonic/gin"
)

func (a *Api) RegisterRoutes(e *gin.Engine) {
	e.POST("/signup", a.SignupUserHandler)
	e.POST("/login", a.LoginUserHandler)
	authenticated := e.Group("/user")
	authenticated.Use(middlewares.Authentication)
	authenticated.POST("/initial", a.GetInitialDataHandler)
	authenticated.GET("/question", a.GetQuestionHandler)
}

func (a *Api) SignupUserHandler(ctx *gin.Context) {
	var u *store.User
	err := ctx.ShouldBindJSON(&u)
	if err != nil {
		fmt.Println("error while binding json", err)
		ctx.JSON(400, gin.H{"success": false, "message": err})
		return
	}
	token, err := a.Store.UserStore.CreateUser(u)
	if err != nil {
		fmt.Println("error while creating document ")
		ctx.JSON(400, gin.H{"success": false, "message": fmt.Sprint(err)})
		return
	}
	ctx.JSON(201, gin.H{"success": true, "token": token})
}

func (a *Api) LoginUserHandler(ctx *gin.Context) {
	var payload *store.LoginPayload
	err := ctx.ShouldBindJSON(&payload)
	if err != nil {
		ctx.JSON(401, gin.H{"success": false, "message": fmt.Sprint(err)})
		return
	}
	//token
	token, err := a.Store.UserStore.LoginUser(payload)
	if err != nil {
		ctx.JSON(401, gin.H{"success": false, "message": fmt.Sprint(err)})
		return
	}
	ctx.JSON(200, gin.H{"success": true, "message": "user logged in successfully", "token": token})
}

func (a *Api) GetInitialDataHandler(ctx *gin.Context) {
	var payload *store.InititalPrompt
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(401, gin.H{"success": false, "message": "incorrect data format"})
		return
	}
	if err := a.Store.ModelStore.GetInitialData(payload); err != nil {
		fmt.Println(err)
		ctx.JSON(500, gin.H{"success": false, "message": "error saving initial data"})
		return
	}
	data, err := a.Store.ModelStore.GetAllH(ctx, payload.UserId)
	if err != nil {
		ctx.JSON(500, gin.H{"success": false, "message": "error fetching data from redis"})
		return
	}
	fmt.Println("Data: ", data)
	ctx.JSON(200, gin.H{"success": true, "message": "initial data recieved", "data": data})
}

func (a *Api) GetQuestionHandler(ctx *gin.Context) {
	userId := ctx.Query("userId")
	fmt.Println("First")
	if userId == "" {
		println(userId)
		ctx.JSON(400, gin.H{"success": false, "message": "invalid userId"})
		return
	}
	correctStr := ctx.DefaultQuery("correctResponses", "0")

	correct, err := strconv.Atoi(correctStr)
	fmt.Println("second")
	if err != nil {
		ctx.JSON(400, gin.H{"success": false, "message": "correctResponses must be a number"})
		return
	}

	payload := &store.Ask{
		UserId:           userId,
		CorrectResponses: correct,
	}

	fmt.Println("third")
	// Setup SSE headers
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	_, err = a.Store.ModelStore.GetQuestion(ctx.Writer, context.Background(), payload)
	if err != nil {
		fmt.Fprintf(ctx.Writer, "event: error\ndata: %s\n\n", err.Error())
		ctx.Writer.Flush()
	}
}
