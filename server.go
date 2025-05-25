package main

import (
	"example/gollama-app/db"
	"example/gollama-app/routes"
	"example/gollama-app/store"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	userDb := db.InitDb()
	server := gin.Default()
	s := store.Store{
		UserCol: userDb,
		Server:  server,
	}
	store.Distribute(&s)
	routes.RegisterRoutes(server)
	err := server.Run(":8080")
	if err != nil {
		fmt.Println(err)
	}
}
