package auth

import (
	"github.com/kazugmx/narubox-bot/internal/auth/pass/argon2"
	db "github.com/kazugmx/narubox-bot/internal/db/sqlc"
)

type AuthHandler struct {
	query     *db.Queries
	argon2sv  *argon2.Config
	dummyHash string
}

type LoginRequestPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegistrationRequestPayload struct {
	Username    string `json:"username"`
	MailAddress string `json:"mailaddr"`
	Password    string `json:"password"`
}
