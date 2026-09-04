package jwt

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kazugmx/narubox-bot/internal/config"
)

type Service struct {
	JWTSecret          []byte
	JWTIssuer          string
	ValidSigningMethod []string
}

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

const claimsKey = "claims"

func NewJWTService(s *config.AppConfig) *Service {
	return &Service{
		JWTIssuer:          s.JWTIssuer,
		JWTSecret:          []byte(s.JWTSecret),
		ValidSigningMethod: []string{jwt.SigningMethodHS256.Alg()},
	}
}

func (s *Service) GenerateToken(userID uuid.UUID) (string, error) {
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

func (s *Service) VerifyToken(tok string) (*Claims, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		return []byte(s.JWTSecret), nil
	}

	token, err := jwt.ParseWithClaims(
		tok,
		&Claims{},
		keyFunc,
		jwt.WithIssuer(s.JWTIssuer),
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

func (s *Service) AuthMiddleware(c fiber.Ctx) error {
	unauthorized := func() error {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "unauthorized",
			"message": "invalid authorization header",
		})
	}
	header := c.Get("Authorization")

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return unauthorized()
	}

	token := strings.TrimPrefix(header, prefix)

	if len(header) == 0 {
		return unauthorized()
	}

	claims, err := s.VerifyToken(token)
	if err != nil {
		return unauthorized()
	}

	c.Locals(claimsKey, claims)
	return c.Next()
}

func GetClaims(c fiber.Ctx) (*Claims, bool) {
	claims, ok := c.Locals("claims").(*Claims)
	return claims, ok
}
