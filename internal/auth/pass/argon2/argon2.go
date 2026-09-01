package argon2

import (
	"github.com/alexedwards/argon2id"
)

type Config struct {
	params *argon2id.Params
}

func NewArgon2Service(params *argon2id.Params) *Config {
	return &Config{
		params: params,
	}
}

func (config *Config) Verify(password string, hash string) (match bool, params *argon2id.Params, err error) {
	return argon2id.CheckHash(password, hash)
}

func (config *Config) GenerateHash(password string) (hash string, err error) {
	return argon2id.CreateHash(password, config.params)
}
