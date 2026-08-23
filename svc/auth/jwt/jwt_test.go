package jwt

import "testing"

func TestNewJWTServiceRequiresConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		origin string
	}{
		{name: "missing secret", origin: "https://example.test"},
		{name: "missing origin", secret: "test-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewJWTService(tt.secret, tt.origin); err == nil {
				t.Fatal("NewJWTService() error = nil, want an error")
			}
		})
	}
}

func TestJWTServiceRoundTrip(t *testing.T) {
	service, err := NewJWTService("test-secret", "https://example.test")
	if err != nil {
		t.Fatalf("NewJWTService() error = %v", err)
	}

	token, err := service.GenerateJwtToken(42)
	if err != nil {
		t.Fatalf("GenerateJwtToken() error = %v", err)
	}
	claims, err := service.VerifyJwtToken(token)
	if err != nil {
		t.Fatalf("VerifyJwtToken() error = %v", err)
	}
	if claims.Subject != "42" {
		t.Fatalf("claims.Subject = %q, want %q", claims.Subject, "42")
	}
}
