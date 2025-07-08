package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Farmaan-Malik/gollama-app/utils"

	"github.com/chenmingyong0423/go-mongox/v2"
	"github.com/chenmingyong0423/go-mongox/v2/builder/query"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserStore struct {
	UserCol *mongox.Collection[User]
}

type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" mongox:"autoID"`
	FirstName string        `bson:"first_name" validate:"required,min=3"`
	LastName  string        `bson:"last_name" validate:"required,min=3"`
	Email     string        `bson:"email" validate:"required,min=8"`
	Username  string        `bson:"username" validate:"required,min=7"`
	Password  string        `bson:"password" validate:"required,min=7"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}
type LoginPayload struct {
	Email    string `json:"email" validate:"required,min=8"`
	Password string `json:"password" validate:"required,min=7"`
}

func (s *UserStore) CreateUser(u *User) (*bson.ObjectID, string, error) {
	exists, err := s.UserCol.Finder().Filter(query.Eq("email", u.Email)).FindOne(context.Background())
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("User Doesnt exist")
		} else {
			return nil, "", err
		}
	}

	if exists != nil {
		fmt.Println("User already exists: ", exists)
		return nil, "", errors.New("user with this email already exists")
	}
	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return nil, "", err
	}
	u.Password = hashedPassword
	result, err := s.UserCol.Creator().InsertOne(context.Background(), u)
	if err != nil {
		return nil, "", err
	}
	id := result.InsertedID.(bson.ObjectID)
	token, err := utils.CreateJwt(utils.JwtPayload{
		UserId:   id.Hex(),
		UserName: u.Username,
	})
	if err != nil {
		return nil, "", err
	}
	return &id, token, nil
}

func (s *UserStore) LoginUser(p *LoginPayload) (*User, string, error) {
	col := s.UserCol
	user, err := col.Finder().Filter(query.Eq("email", p.Email)).FindOne(context.Background())
	if err != nil {
		return nil, "", errors.New("incorrect user/password")
	}
	err = utils.CompareHash(p.Password, user.Password)
	if err != nil {
		return nil, "", err
	}
	token, err := utils.CreateJwt(utils.JwtPayload{
		UserId: user.ID.Hex(),
	})
	return user, token, nil
}
