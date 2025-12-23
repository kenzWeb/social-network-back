package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	UserID   string `json:"sub"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID string `json:"sub"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}


func (c AccessClaims) Validate() error {
	if c.Type != "access" {
		return errors.New("invalid token type: expected access")
	}
	if c.UserID == "" {
		return errors.New("missing subject (user ID)")
	}
	return nil
}

func (c RefreshClaims) Validate() error {
	if c.Type != "refresh" {
		return errors.New("invalid token type: expected refresh")
	}
	if c.UserID == "" {
		return errors.New("missing subject (user ID)")
	}
	return nil
}
