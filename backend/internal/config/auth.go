package config

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	authUsernameEnv     = "CMS_AUTH_USERNAME"
	authPasswordHashEnv = "CMS_AUTH_PASSWORD_HASH"
	authJWTSecretEnv    = "CMS_AUTH_JWT_SECRET"
	minJWTSecretBytes   = 32
)

type EnvLookup func(string) (string, bool)

func ApplyAuthEnv(c *Config, lookup EnvLookup) error {
	if c == nil || lookup == nil {
		return errors.New("invalid configuration loader")
	}

	// Credentials are environment-only and never fall back to repository YAML files.
	c.Auth.Username = ""
	c.Auth.PasswordHash = ""
	c.Auth.JwtSecret = ""

	username, ok := requiredEnv(lookup, authUsernameEnv)
	if !ok {
		return fmt.Errorf("%s is required", authUsernameEnv)
	}
	if len(username) > 128 {
		return fmt.Errorf("%s must not exceed 128 characters", authUsernameEnv)
	}
	if containsInsecurePlaceholder(username) {
		return fmt.Errorf("%s contains an insecure placeholder", authUsernameEnv)
	}

	passwordHash, ok := requiredEnv(lookup, authPasswordHashEnv)
	if !ok {
		return fmt.Errorf("%s is required", authPasswordHashEnv)
	}
	cost, err := bcrypt.Cost([]byte(passwordHash))
	if err != nil {
		return fmt.Errorf("%s must be a valid bcrypt hash", authPasswordHashEnv)
	}
	if cost < bcrypt.DefaultCost {
		return fmt.Errorf("%s bcrypt cost must be at least %d", authPasswordHashEnv, bcrypt.DefaultCost)
	}

	jwtSecret, ok := requiredEnv(lookup, authJWTSecretEnv)
	if !ok {
		return fmt.Errorf("%s is required", authJWTSecretEnv)
	}
	if len(jwtSecret) < minJWTSecretBytes {
		return fmt.Errorf("%s must contain at least %d bytes", authJWTSecretEnv, minJWTSecretBytes)
	}
	if containsInsecurePlaceholder(jwtSecret) {
		return fmt.Errorf("%s contains an insecure placeholder", authJWTSecretEnv)
	}
	if c.Auth.JwtExpireSec <= 0 {
		return errors.New("Auth.JwtExpireSec must be greater than zero")
	}

	c.Auth.Username = username
	c.Auth.PasswordHash = passwordHash
	c.Auth.JwtSecret = jwtSecret
	return nil
}

func containsInsecurePlaceholder(value string) bool {
	lowerValue := strings.ToLower(value)
	for _, marker := range []string{"replace-me", "replace-with", "change-me", "example"} {
		if strings.Contains(lowerValue, marker) {
			return true
		}
	}
	return false
}

func requiredEnv(lookup EnvLookup, key string) (string, bool) {
	value, ok := lookup(key)
	value = strings.TrimSpace(value)
	return value, ok && value != ""
}
