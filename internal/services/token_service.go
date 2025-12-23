package services

import (
	"errors"
	"time"

	"modern-social-media/internal/auth"
	"modern-social-media/internal/models"

	"github.com/golang-jwt/jwt/v5"
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
	now := time.Now()
	claims := auth.AccessClaims{
		UserID:   u.ID,
		Email:    u.Email,
		Username: u.Username,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.AccessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.Secret)
}

func (s *JWTTokenService) IssueRefresh(u *models.User) (string, error) {
	now := time.Now()
	claims := auth.RefreshClaims{
		UserID: u.ID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.RefreshTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.Secret)
}

func (s *JWTTokenService) ParseRefresh(tokenStr string) (string, error) {
	var claims auth.RefreshClaims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.Secret, nil
	})

	if err != nil {
		return "", errors.New("invalid_refresh")
	}

	if !token.Valid {
		return "", errors.New("invalid_refresh")
	}

	if err := claims.Validate(); err != nil {
		return "", errors.New("invalid_refresh")
	}

	return claims.UserID, nil
}
