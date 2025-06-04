package db

import (
	"fmt"
	"os"

	"github.com/Farmaan-Malik/gollama-app/store"

	"github.com/chenmingyong0423/go-mongox/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitDb() *mongox.Collection[store.User] {
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	panic(err)
	// }
	uri, ok := os.LookupEnv("MONGO_URI")
	if !ok {
		panic("mongo_uri not found in environment variables")
	}
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
