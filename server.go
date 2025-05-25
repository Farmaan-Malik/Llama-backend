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
