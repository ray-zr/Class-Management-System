package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func Sign(username string, secret string, expireSec int64, now time.Time) (token string, expireAt int64, err error) {
	if secret == "" {
		return "", 0, errors.New("empty jwt secret")
	}
	exp := now.Add(time.Duration(expireSec) * time.Second)
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	jt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := jt.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return str, exp.Unix(), nil
}

func Parse(tokenString string, secret string) (*Claims, error) {
	if secret == "" {
		return nil, errors.New("empty jwt secret")
	}
	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
