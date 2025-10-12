package services

import (
	"modern-social-media/internal/auth"
)

type PasswordHasher interface {
	Hash(pwd string) (string, error)
	Verify(hash, pwd string) (bool, error)
}

type Argon2Hasher struct{}

func (Argon2Hasher) Hash(pwd string) (string, error)       { return auth.HashPassword(pwd) }
func (Argon2Hasher) Verify(hash, pwd string) (bool, error) { return auth.VerifyPassword(hash, pwd) }
