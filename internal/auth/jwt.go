package jwtOperator

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret   []byte
	issuer   string
	origin   string
	audience string
}

func NewJWTService(secret string) *JWTService {
	origin := os.Getenv("CALLBACK_ORIGIN")
	if !(len(origin) > 0) {
		log.Fatalln("error\t env CALLBACK_ORIGIN is not set.")
	}

	return &JWTService{
		secret:   []byte(secret),
		issuer:   "narubox-auth/",
		origin:   "" + "bd",
		audience: "narubox-bot webclient/api",
	}
}

func (j JWTService) GenerateJwtToken(userId int32) (res_token string, err error) {
	//これどう実装したらいいんだろ
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId,
		"iss": j.issuer,
		"aud": j.audience,
		"iat": time.Now().Unix(),
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
		return false, fmt.Errorf("Invalid token")
	} else if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if time.Now().Unix() > int64(claims["exp"].(float64)) {
			// expire check
			return false, fmt.Errorf("Expired token.")
		} else {

			return true, nil
		}
	} else {
		return false, fmt.Errorf("Failed to parse")
	}
}
