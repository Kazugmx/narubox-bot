package Auth

import (
	"github.com/Kazugmx/narubox-bot/db"
	jwtOperator "github.com/Kazugmx/narubox-bot/svc/auth/jwt"
)

// Handler
type AuthHandler struct {
	jwtEngine *jwtOperator.JWTService
	query     *db.Queries
}

// Payloads
type UserLoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SelfStatusRes struct {
	Mailaddr   string  `json:"mail"`
	Username   string  `json:"username"`
	CreatedAt  *string `json:"createdAt,omitempty"`
	LastAccess *string `json:"lastAccess,omitempty"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
