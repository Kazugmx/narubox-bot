package Auth

import (
	"time"
)

// UserRegistration & Login payloads

type UserCreatePayload struct {
	Username string `json:"username"`
	Mailaddr string `json:"mail"`
	Password string `json:"password"`
}

type UserCreateRes struct {
	Status   string `json:"status"`
	DateTime *time.Time
}

type UserLoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Self information
type SelfStatusRes struct {
	Mailaddr   string     `json:"mail"`
	Username   string     `json:"username"`
	CreatedAt  *time.Time `json:"createdAt"`
	LastAccess *time.Time `json:"lastAccess"`
}
