package db

import (
	"context"
	"example/gollama-app/store"
	"fmt"
	"os"

	"github.com/chenmingyong0423/go-mongox/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func InitDb() *mongox.Collection[store.User] {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	uri := os.Getenv("MONGO_URI")
	fmt.Println(uri)
	c, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	client := mongox.NewClient(c, &mongox.Config{})
	db := client.NewDatabase("user")
	userCol := mongox.NewCollection[store.User](db, "users_profile")
	if err := client.Client().Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return userCol
}
