package jwtOperator

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: []byte(secret),
	}
}

func (j JWTService) GenerateJwtToken(userId int32) (res_token string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenStr, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("JWT Signing failed: %v", err)
	}

	return tokenStr, nil
}

func (j JWTService) VerifyJwtToken(chall string) (bool, error) {
	token, err := jwt.Parse(chall, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		log.Printf("invalid token. error: %v", err)
	} else if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if time.Now().Unix() > int64(claims["exp"].(float64)) {
			return false, fmt.Errorf("Expired token.")
		} else if false { //placeholder

		} else {
			return true, nil
		}
	} else {
		return false, fmt.Errorf("Failed to parse")
	}
}
