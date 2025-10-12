package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"modern-social-media/internal/models"
	"modern-social-media/internal/utils"
)

type UserRepo interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	CreateUser(ctx context.Context, u *models.User) error
	UpdateUser(ctx context.Context, u *models.User) error
}

type CodeRepo interface {
	Create(ctx context.Context, v *models.VerificationCode) error
	GetValid(ctx context.Context, userID, purpose, code string) (*models.VerificationCode, error)
	Consume(ctx context.Context, id string) error
	DeleteByUserAndPurpose(ctx context.Context, userID, purpose string) error
	DeleteExpired(ctx context.Context) error
}

type Mailer interface {
	Send(to, subject, body string) error
}

type TxRunner interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type AuthService struct {
	Users           UserRepo
	Codes           CodeRepo
	Tokens          TokenService
	Mailer          Mailer
	Hasher          PasswordHasher
	Clock           Clock
	Transact        TxRunner
	Email2FAEnabled bool
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*models.User, error) {
	username := strings.TrimSpace(in.Username)
	email := utils.NormalizeEmail(in.Email)
	password := strings.TrimSpace(in.Password)
	first := strings.TrimSpace(in.FirstName)
	last := strings.TrimSpace(in.LastName)
	if username == "" || email == "" || password == "" {
		return nil, errors.New("missing_required_fields")
	}
	if !utils.IsValidEmail(email) {
		return nil, errors.New("invalid_email")
	}
	if len(password) < 8 {
		return nil, errors.New("weak_password")
	}
	hash, err := s.Hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	var created *models.User
	err = s.Transact.WithTx(ctx, func(ctx context.Context) error {
		existingEmail, _ := s.Users.GetByEmail(ctx, email)
		existingUser, _ := s.Users.GetByUsername(ctx, username)

		reuse := func(u *models.User) (*models.User, error) {
			u.Username = username
			u.Email = email
			u.Password = hash
			u.FirstName = first
			u.LastName = last
			u.IsVerified = false
			u.IsActive = true
			if err := s.Users.UpdateUser(ctx, u); err != nil {
				return nil, err
			}
			if err := s.Codes.DeleteByUserAndPurpose(ctx, u.ID, "email_verify"); err != nil {
				return nil, err
			}
			return u, nil
		}

		if existingEmail != nil {
			if existingEmail.IsVerified {
				return errors.New("email_in_use")
			}
			if existingUser != nil && existingUser.ID != existingEmail.ID {
				return errors.New("username_in_use")
			}
			u, err := reuse(existingEmail)
			if err != nil {
				return err
			}
			created = u
		} else if existingUser != nil {
			if existingUser.IsVerified || strings.ToLower(existingUser.Email) != email {
				return errors.New("username_in_use")
			}
			u, err := reuse(existingUser)
			if err != nil {
				return err
			}
			created = u
		} else {
			u := &models.User{Username: username, Email: email, Password: hash, FirstName: first, LastName: last}
			if err := s.Users.CreateUser(ctx, u); err != nil {
				return err
			}
			created = u
		}

		if err := s.Codes.DeleteByUserAndPurpose(ctx, created.ID, "email_verify"); err != nil {
			return err
		}
		code := utils.GenerateDigits(6)
		v := &models.VerificationCode{UserID: created.ID, Purpose: "email_verify", Code: code, ExpiresAt: s.Clock.Now().Add(30 * time.Minute)}
		if err := s.Codes.Create(ctx, v); err != nil {
			return err
		}
		_ = s.Mailer.Send(created.Email, "Подтверждение почты", "Ваш код подтверждения: "+code)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, email, code string) error {
	email = utils.NormalizeEmail(email)
	code = strings.TrimSpace(code)
	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("not_found")
	}
	v, err := s.Codes.GetValid(ctx, u.ID, "email_verify", code)
	if err != nil {
		return errors.New("invalid_code")
	}
	if err := s.Codes.Consume(ctx, v.ID); err != nil {
		return err
	}
	u.IsVerified = true
	return s.Users.UpdateUser(ctx, u)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthTokens, error) {
	email = utils.NormalizeEmail(email)
	password = strings.TrimSpace(password)
	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid_credentials")
	}
	ok, err := s.Hasher.Verify(u.Password, password)
	if err != nil || !ok {
		return nil, errors.New("invalid_credentials")
	}
	if !u.IsVerified {
		return nil, errors.New("email_not_verified")
	}
	if s.Email2FAEnabled && u.Is2FAEnabled {
		return nil, errors.New("2fa_required")
	}
	access, err := s.Tokens.IssueAccess(u)
	if err != nil {
		return nil, err
	}
	refresh, err := s.Tokens.IssueRefresh(u)
	if err != nil {
		return nil, err
	}
	return &AuthTokens{Access: access, Refresh: refresh}, nil
}
