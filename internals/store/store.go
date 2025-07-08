package store

import (
	"context"
	"net/http"

	"github.com/chenmingyong0423/go-mongox/v2"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	UserStore interface {
		CreateUser(u *User) (string, error)
		LoginUser(p *LoginPayload) (string, error)
	}
	ModelStore interface {
		GetQuestion(w http.ResponseWriter, ctx context.Context, a *Ask) (*Question, error)
		GetInitialData(i *InititalPrompt) error
		GetAllH(ctx context.Context, key string) (map[string]string, error)
	}
}

func NewStore(r *redis.Client, mongoCol *mongox.Collection[User]) *Store {
	return &Store{
		UserStore: &UserStore{
			UserCol: mongoCol,
		},
		ModelStore: &ModelStore{
			Redis: r,
		},
	}
}
