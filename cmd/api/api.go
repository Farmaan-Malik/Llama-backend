package main

import (
	"github.com/Farmaan-Malik/gollama-app/internals/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Api struct {
	Store *store.Store
}

func NewApi(s *store.Store) *Api {
	return &Api{
		Store: s,
	}
}

func (a *Api) listenAndServe() error {
	server := gin.Default()
	server.Use(cors.Default()) // All origins allowed by default
	a.RegisterRoutes(server)
	return server.Run(":8080")

}
