package config

type AppConfig struct {
	DatabaseURL    string
	JWTSecret      string
	JWTIssuer      string
	CallbackOrigin string
	URIMaster      string
	APIKey         string
}
