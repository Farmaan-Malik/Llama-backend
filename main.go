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
		Redis:   redis,
	}
	defer redis.Close()
	store.Distribute(&s)
	var api routes.Api = routes.Api{
		Store: &s,
	}
	api.RegisterRoutes(server)
	// cmd := s.Redis.HGet(context.Background(), "12", "subject")
	// fmt.Println(cmd)
	err := server.Run(":8080")
	if err != nil {
		fmt.Println(err)
	}
}
