package main

import (
	"fmt"

	"github.com/Farmaan-Malik/gollama-app/db"
	"github.com/Farmaan-Malik/gollama-app/routes"
	"github.com/Farmaan-Malik/gollama-app/store"

	"github.com/gin-gonic/gin"
)

func main() {
	userDb := db.InitDb()
	redis := db.InitRedis()
	server := gin.Default()
	s := store.Store{
		UserCol: userDb,
		Server:  server,
		Redis:   redis,
	}
	defer redis.Close()
	store.Distribute(&s)
	routes.RegisterRoutes(server)
	var ask store.Ask = store.Ask{
		QuestionsAsked:   []string{},
		Subject:          "History",
		Standard:         "6",
		CorrectResponses: 0,
		UserId:           "12",
	}
ask.GetQuestion()
	err := server.Run(":8080")
	if err != nil {
		fmt.Println(err)
	}

}
