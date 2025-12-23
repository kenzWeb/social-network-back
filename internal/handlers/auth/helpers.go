package auth

import (
	"errors"
	"time"

	"modern-social-media/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func setRefreshCookie(c *gin.Context, token string) {
	c.SetCookie(
		"refreshToken",
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
	var claims auth.AccessClaims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", errors.New("invalid token")
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	if err := claims.Validate(); err != nil {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}
