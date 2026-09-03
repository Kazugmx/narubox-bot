package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type service struct {
	JWTSecret          []byte
	JWTIssuer          string
	ValidSigningMethod []string
}

type ServiceConfig struct {
	JWTSecret string
	JWTIssuer string
}

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

func NewJWTService(s *ServiceConfig) *service {
	return &service{
		JWTIssuer:          s.JWTIssuer,
		JWTSecret:          []byte(s.JWTSecret),
		ValidSigningMethod: []string{jwt.SigningMethodHS256.Alg()},
	}
}

func (s *service) GenerateToken(userID uuid.UUID) (string, error) {
	currentTime := time.Now()
	claims := Claims{
		UserID:    userID,
		Subject:   userID.String(),
		Issuer:    s.JWTIssuer,
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(30 * time.Minute)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.JWTSecret))
}

func (s *service) VerifyToken(tok string) (*Claims, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		return []byte(s.JWTSecret), nil
	}

	token, err := jwt.ParseWithClaims(
		tok,
		&Claims{},
		keyFunc,
		jwt.WithIssuer(string(s.JWTIssuer)),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods(s.ValidSigningMethod),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid claims type")
	}

	return claims, nil
}
