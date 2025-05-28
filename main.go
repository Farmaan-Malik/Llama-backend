package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Farmaan-Malik/gollama-app/db"
	"github.com/Farmaan-Malik/gollama-app/routes"
	"github.com/Farmaan-Malik/gollama-app/store"
	"golang.org/x/net/context"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()
	userDb := db.InitDb()
	redis := db.InitRedis()
	server := gin.Default()
	s := store.Store{
		UserCol: userDb,
		Redis:   redis,
	}
	defer redis.Close()
	store.Distribute(&s)
	var api routes.Api = routes.Api{
		Store: &s,
	}
	api.RegisterRoutes(server)
	var initialPrompt store.InititalPrompt = store.InititalPrompt{
		UserId:   "12",
		Standard: "9",
		Subject:  "English",
	}
	err := s.GetInitialData(&initialPrompt)
	if err != nil {
		log.Fatal(err)
	}
	var ask store.Ask = store.Ask{
		CorrectResponses: 0,
		UserId:           "12",
	}

	question, err := s.GetQuestion(ctx, &ask)
	if err != nil {
		fmt.Println(err)
	}
	jsonBytes, err := json.MarshalIndent(question, "", "  ")
	if err != nil {
		log.Fatalf("error marshaling question to JSON: %v", err)
	}
	fmt.Println("Question: ", string(jsonBytes))

	err = os.WriteFile("response.json", jsonBytes, 0644)
	if err != nil {
		log.Fatalf("error writing to file: %v", err)
	}
	// cmd := s.Redis.HGet(context.Background(), "12", "subject")
	// fmt.Println(cmd)
	err = server.Run(":8080")
	if err != nil {
		fmt.Println(err)
	}

}
