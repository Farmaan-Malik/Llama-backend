package main

import (
	"github.com/Farmaan-Malik/gollama-app/internals/store"
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

	a.RegisterRoutes(server)
	return server.Run(":8080")

}
