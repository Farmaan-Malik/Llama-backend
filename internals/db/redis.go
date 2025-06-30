package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type UserSession struct {
	UserId         int
	AskedQuestions []string
}

func InitRedis(redisPassword string, addr string) *redis.Client {
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
