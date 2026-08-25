package config

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestApplyAuthEnv(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("a-secure-test-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	env := map[string]string{
		authUsernameEnv:     "teacher-admin",
		authPasswordHashEnv: string(hash),
		authJWTSecretEnv:    strings.Repeat("s", minJWTSecretBytes),
	}
	c := Config{}
	c.Auth.Username = "yaml-user-must-not-be-used"
	c.Auth.PasswordHash = "yaml-hash-must-not-be-used"
	c.Auth.JwtSecret = "yaml-secret-must-not-be-used"
	c.Auth.JwtExpireSec = 3600

	if err := ApplyAuthEnv(&c, mapLookup(env)); err != nil {
		t.Fatalf("ApplyAuthEnv() error = %v", err)
	}
	if c.Auth.Username != env[authUsernameEnv] || c.Auth.PasswordHash != env[authPasswordHashEnv] || c.Auth.JwtSecret != env[authJWTSecretEnv] {
		t.Fatal("ApplyAuthEnv() did not apply environment credentials")
	}
}

func TestApplyAuthEnvRejectsInvalidConfiguration(t *testing.T) {
	validHash, err := bcrypt.GenerateFromPassword([]byte("a-secure-test-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	validEnv := map[string]string{
		authUsernameEnv:     "teacher-admin",
		authPasswordHashEnv: string(validHash),
		authJWTSecretEnv:    strings.Repeat("s", minJWTSecretBytes),
	}
	tests := []struct {
		name      string
		mutateEnv func(map[string]string)
		expireSec int64
	}{
		{name: "missing username", mutateEnv: func(env map[string]string) { delete(env, authUsernameEnv) }, expireSec: 3600},
		{name: "placeholder username", mutateEnv: func(env map[string]string) { env[authUsernameEnv] = "replace-with-admin-username" }, expireSec: 3600},
		{name: "invalid password hash", mutateEnv: func(env map[string]string) { env[authPasswordHashEnv] = "not-bcrypt" }, expireSec: 3600},
		{name: "short JWT secret", mutateEnv: func(env map[string]string) { env[authJWTSecretEnv] = "too-short" }, expireSec: 3600},
		{name: "placeholder JWT secret", mutateEnv: func(env map[string]string) { env[authJWTSecretEnv] = "replace-me-with-a-long-secret-value" }, expireSec: 3600},
		{name: "example JWT secret", mutateEnv: func(env map[string]string) { env[authJWTSecretEnv] = "replace-with-at-least-32-random-bytes" }, expireSec: 3600},
		{name: "invalid expiry", mutateEnv: func(map[string]string) {}, expireSec: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := cloneMap(validEnv)
			tt.mutateEnv(env)
			c := Config{}
			c.Auth.JwtExpireSec = tt.expireSec
			if err := ApplyAuthEnv(&c, mapLookup(env)); err == nil {
				t.Fatal("ApplyAuthEnv() accepted invalid authentication configuration")
			}
			if c.Auth.Username != "" || c.Auth.PasswordHash != "" || c.Auth.JwtSecret != "" {
				t.Fatal("ApplyAuthEnv() retained partial credentials after validation failure")
			}
		})
	}
}

func mapLookup(values map[string]string) EnvLookup {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func cloneMap(values map[string]string) map[string]string {
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
