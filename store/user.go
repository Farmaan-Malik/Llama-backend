package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Farmaan-Malik/gollama-app/utils"

	"github.com/chenmingyong0423/go-mongox/v2/builder/query"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" mongox:"autoID"`
	FirstName string        `bson:"first_name"`
	LastName  string        `bson:"last_name"`
	Email     string        `bson:"email"`
	Username  string        `bson:"username"`
	Password  string        `bson:"password"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}
type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Store) CreateUser(u *User) (*mongo.InsertOneResult, error) {

	exists, err := s.UserCol.Finder().Filter(query.Eq("email", u.Email)).FindOne(context.Background())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("User Doesnt exist")
		} else {
			return nil, err
		}
	}
	if exists != nil {
		fmt.Println("User already exists")
		return nil, errors.New("user with this email already exists")
	}
	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return nil, err
	}
	u.Password = hashedPassword
	result, err := s.UserCol.Creator().InsertOne(context.Background(), u)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) LoginUser(p *LoginPayload) (bool, error) {
	col := Data.UserCol
	exists, err := col.Finder().Filter(query.Eq("email", p.Email)).FindOne(context.Background())
	if err != nil {
		return false, errors.New("incorrect user/password")
	}
	err = utils.CompareHash(p.Password, exists.Password)
	if err != nil {
		return false, err
	}
	return true, nil
}
