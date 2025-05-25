package store

import (
	"github.com/chenmingyong0423/go-mongox/v2"
	"github.com/gin-gonic/gin"
)

type Store struct {
	UserCol *mongox.Collection[User]
	Server  *gin.Engine
}

var Data *Store

func Distribute(s *Store) {
	Data = s
}
