package utils

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type JwtPayload struct {
	UserId   string
	UserName string
}

func CreateJwt(j JwtPayload) (string, error) {
	key, ok := os.LookupEnv("JWT_KEY")
	if !ok {
		key = "fallback"
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":    "GoLlama",
		"userId": j.UserId,
		"name":   j.UserName,
	})
	s, err := t.SignedString([]byte(key))
	if err != nil {
		return "", err
	}
	return s, nil
}

func ValidateJwt(s string) error {
	ParsedToken, err := jwt.Parse(s, func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("invalid signing method")
		}
		key, ok := os.LookupEnv("JWT_KEY")
		if !ok {
			key = "fallback"
		}
		return []byte(key), nil
	})
	if err != nil {
		return err
	}
	if !ParsedToken.Valid {
		return errors.New("invalid token")
	}
	claims, valid := ParsedToken.Claims.(jwt.MapClaims)
	if !valid {
		return errors.New("invalid token claims")
	}
	issuer, err := claims.GetIssuer()
	if err != nil {
		return err
	}
	if issuer != "GoLlama" {
		return errors.New("invalid issuer")
	}

	return nil
}
