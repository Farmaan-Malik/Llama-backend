package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Farmaan-Malik/gollama-app/internals/db"
	"github.com/Farmaan-Malik/gollama-app/internals/store"
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func main() {
	// Uncomment if using server outside of docker
	// This would load .env file's content to your environment variables
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	// Use environment variables
	// If running outside the container without loading .env, this would fail.

	mongoURI := os.Getenv("MONGO_URI")
	fmt.Println("Mongo URI:", mongoURI)
	// redisAddr := os.Getenv("REDIS_ADDR")
	// redisPw := os.Getenv("REDIS_PW")
	// fmt.Println("Redis Addr:", redisAddr)
	userDb := db.InitDb(mongoURI)

	// Uncomment if using redis from cloud
	// redis := db.InitRedis(redisPw, redisAddr)

	redis := db.InitRedis("", "redis:6379")
	s := store.NewStore(redis, userDb)

	defer redis.Close()
	api := NewApi(s)
	if err := api.listenAndServe(); err != nil {
		log.Fatal("error starting server ", err)
	}
}
