package handlers

import (
	"errors"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func Refresh(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("refresh_token")
		if err != nil || cookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
			return
		}
		parsed, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !parsed.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}
		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok || claims["type"] != "refresh" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
			return
		}
		sub, _ := claims["sub"].(string)
		accClaims := jwt.MapClaims{
			"sub":  sub,
			"type": "access",
			"exp":  time.Now().Add(15 * time.Minute).Unix(),
			"iat":  time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, accClaims)
		signed, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": signed})
	}
}
