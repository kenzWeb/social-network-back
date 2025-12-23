package middleware

import (
	"errors"
	"net/http"
	"strings"

	"modern-social-media/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			return
		}

		claims, err := validateAccessToken(token, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("isAdmin", claims.IsAdmin)
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	authz := c.GetHeader("Authorization")
	if !strings.HasPrefix(authz, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authz, "Bearer ")
}

func validateAccessToken(tokenStr, secret string) (*auth.AccessClaims, error) {
	var claims auth.AccessClaims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, errors.New("invalid token")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if err := claims.Validate(); err != nil {
		return nil, err
	}

	return &claims, nil
}

func StaticToken(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if required == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "forbidden"})
			return
		}
		token := c.GetHeader("X-Admin-Token")
		if token == "" || token != required {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
