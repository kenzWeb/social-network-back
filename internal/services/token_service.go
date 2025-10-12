package services

import (
	"errors"
	"time"

	"modern-social-media/internal/models"

	jwt "github.com/dgrijalva/jwt-go"
)

type TokenService interface {
	IssueAccess(u *models.User) (string, error)
	IssueRefresh(u *models.User) (string, error)
	ParseRefresh(token string) (string, error)
}

type JWTTokenService struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func (s *JWTTokenService) IssueAccess(u *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      u.ID,
		"email":    u.Email,
		"username": u.Username,
		"type":     "access",
		"exp":      time.Now().Add(s.AccessTTL).Unix(),
		"iat":      time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.Secret)
}

func (s *JWTTokenService) IssueRefresh(u *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"type": "refresh",
		"exp":  time.Now().Add(s.RefreshTTL).Unix(),
		"iat":  time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.Secret)
}

func (s *JWTTokenService) ParseRefresh(tokenStr string) (string, error) {
	p, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return s.Secret, nil })
	if err != nil || !p.Valid {
		return "", errors.New("invalid_refresh")
	}
	c, ok := p.Claims.(jwt.MapClaims)
	if !ok || c["type"] != "refresh" {
		return "", errors.New("invalid_refresh")
	}
	sub, _ := c["sub"].(string)
	return sub, nil
}
