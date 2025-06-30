package db

import (
	"fmt"

	"github.com/Farmaan-Malik/gollama-app/internals/store"

	"github.com/chenmingyong0423/go-mongox/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitDb(uri string) *mongox.Collection[store.User] {

	fmt.Println(uri)
	c, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	client := mongox.NewClient(c, &mongox.Config{})
	db := client.NewDatabase("user")
	userCol := mongox.NewCollection[store.User](db, "users_profile")
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return userCol
}
