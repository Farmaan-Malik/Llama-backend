package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type UserSession struct {
	UserId         int
	AskedQuestions []string
}

func InitRedis() *redis.Client {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	redisPassword, ok := os.LookupEnv("REDIS_PW")
	if !ok {
		log.Fatal("environment variable not found (redis_pw)")
	}

	addr, ok := os.LookupEnv("REDIS_ADDR")
	if !ok {
		log.Fatal("environment variable not found (redis_addr)")
	}
	fmt.Println(redisPassword)
	fmt.Println(addr)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: redisPassword,
		DB:       0,
		Protocol: 2,
	})
	status := client.Ping(context.Background())
	fmt.Println(status)
	return client
}
