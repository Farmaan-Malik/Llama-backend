package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Farmaan-Malik/gollama-app/internals/db"
	"github.com/Farmaan-Malik/gollama-app/internals/store"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Use environment variables
	mongoURI := os.Getenv("MONGO_URI")
	fmt.Println("Mongo URI:", mongoURI)
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPw := os.Getenv("REDIS_PW")
	fmt.Println("Redis Addr:", redisAddr)
	userDb := db.InitDb(mongoURI)
	redis := db.InitRedis(redisPw, redisAddr)

	s := store.NewStore(redis, userDb)

	defer redis.Close()
	api := NewApi(s)
	if err := api.listenAndServe(); err != nil {
		log.Fatal("error starting server ", err)
	}
}
