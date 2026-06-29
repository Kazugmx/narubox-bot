package jwtOperator

import (
	"fmt"
	"log"
	"log/slog"
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
		issuer:   fmt.Sprintf("narubox-auth/%s", origin),
		origin:   origin,
		audience: "narubox-bot webclient/api",
	}
}

func (j JWTService) GenerateJwtToken(userId int32) (res_token string, err error) {
	//キチゲ発散モデル
	// なにこのキチゲ発散モデルってメモは...
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		Issuer:    j.issuer,
		Audience:  jwt.ClaimStrings{j.audience},
		Subject:   fmt.Sprintf("%d", userId),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	})
	slog.Info("testtoken", slog.Any("ExpiresAt", token))

	tokenStr, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("JWT Signing failed: %v", err)
	}

	return tokenStr, nil
}

func (j JWTService) VerifyJwtToken(chall string) (bool, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		chall, claims,
		func(token *jwt.Token) (interface{}, error) {
			return j.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithAllAudiences(j.audience),
		jwt.WithIssuer(j.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithNotBeforeRequired(),
	)

	if err != nil {
		slog.Error("error parsing JWT token:", slog.Any("error", err))
		return false, fmt.Errorf("invalid token")
	}
	if !token.Valid {
		return false, fmt.Errorf("JWT Token is invalid")
	}

	return true, nil
}
