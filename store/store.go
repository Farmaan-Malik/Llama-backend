package store

import (
	"github.com/chenmingyong0423/go-mongox/v2"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	UserCol *mongox.Collection[User]
	Redis   *redis.Client
}

var Data *Store

func Distribute(s *Store) {
	Data = s
}
