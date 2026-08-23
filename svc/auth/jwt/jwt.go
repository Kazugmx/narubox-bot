package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret   []byte
	issuer   string
	audience string
}

func NewJWTService(secret, origin string) (*JWTService, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	if origin == "" {
		return nil, errors.New("CALLBACK_ORIGIN is not set")
	}

	return &JWTService{
		secret:   []byte(secret),
		issuer:   fmt.Sprintf("narubox-auth/%s", origin),
		audience: "narubox-bot webclient/api",
	}, nil
}

func (j JWTService) GenerateJwtToken(userID int32) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		Issuer:    j.issuer,
		Audience:  jwt.ClaimStrings{j.audience},
		Subject:   fmt.Sprintf("%d", userID),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	})

	tokenStr, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("JWT signing failed: %w", err)
	}

	return tokenStr, nil
}

func (j JWTService) VerifyJwtToken(chall string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		chall,
		claims,
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
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("JWT token is invalid")
	}
	return claims, nil
}
