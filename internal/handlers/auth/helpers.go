package auth

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func setRefreshCookie(c *gin.Context, token string) {
	c.SetCookie(
		"refresh_token",
		token,
		int((14 * 24 * time.Hour).Seconds()),
		"/",
		"",
		true,
		true,
	)
}

func parseAccessSubject(jwtSecret string, authz string) (string, error) {
	if len(authz) < 8 || authz[:7] != "Bearer " {
		return "", errors.New("missing bearer token")
	}
	tokenStr := authz[7:]
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !parsed.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "access" {
		return "", errors.New("invalid token")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", errors.New("invalid token")
	}
	return sub, nil
}
