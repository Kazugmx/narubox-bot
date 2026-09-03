package bcrypt

import (
	"golang.org/x/crypto/bcrypt"
)

type Result struct {
	Correct bool
	E       error
}

func CheckPasswordHash(password string, hash string) Result {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return Result{
		err == nil,
		err,
	}
}
