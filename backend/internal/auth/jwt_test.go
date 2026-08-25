package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestSignAndParse(t *testing.T) {
	now := time.Now().Add(-time.Minute)
	token, expireAt, err := Sign("teacher", "test-secret", 3600, now)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if expireAt != now.Add(time.Hour).Unix() {
		t.Fatalf("expireAt = %d, want %d", expireAt, now.Add(time.Hour).Unix())
	}
	claims, err := Parse(token, "test-secret")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.Username != "teacher" {
		t.Fatalf("username = %q, want teacher", claims.Username)
	}
}

func TestParseRejectsUnexpectedAlgorithm(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
		Username: "teacher",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	if _, err := Parse(signed, "test-secret"); err == nil {
		t.Fatal("Parse() accepted a non-HS256 token")
	}
}

func TestSignRejectsEmptySecret(t *testing.T) {
	if _, _, err := Sign("teacher", "", 3600, time.Now()); err == nil {
		t.Fatal("Sign() accepted an empty secret")
	}
}
