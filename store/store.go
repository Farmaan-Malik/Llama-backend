package store

import (
	"github.com/chenmingyong0423/go-mongox/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	UserCol *mongox.Collection[User]
	Server  *gin.Engine
	Redis   *redis.Client
}

var Data *Store

func Distribute(s *Store) {
	Data = s
}
