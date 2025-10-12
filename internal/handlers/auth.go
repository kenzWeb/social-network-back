package handlers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	netmail "net/mail"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"modern-social-media/internal/auth"
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
	"modern-social-media/internal/services"
)

func Register(usersRepo repository.UserRepository, codesRepo repository.VerificationCodeRepository, mail services.EmailSender, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username  string `json:"username"`
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		req.Email = strings.TrimSpace(req.Email)
		req.Password = strings.TrimSpace(req.Password)
		req.FirstName = strings.TrimSpace(req.FirstName)
		req.LastName = strings.TrimSpace(req.LastName)

		normalizedEmail := strings.ToLower(req.Email)

		if req.Username == "" || normalizedEmail == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
			return
		}
		if !isValidEmail(normalizedEmail) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
			return
		}
		if len(req.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
			return
		}

		ctx := c.Request.Context()

		existingByEmail, err := usersRepo.GetByEmail(ctx, normalizedEmail)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				existingByEmail = nil
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email"})
				return
			}
		}

		var existingByUsername *models.User
		if u, err := usersRepo.GetByUsername(ctx, req.Username); err == nil {
			existingByUsername = u
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username"})
			return
		}

		hashed, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		reusePending := func(target *models.User) (*models.User, error) {
			target.Username = req.Username
			target.Email = normalizedEmail
			target.FirstName = req.FirstName
			target.LastName = req.LastName
			target.Password = hashed
			target.IsVerified = false
			target.IsActive = true
			if err := usersRepo.UpdateUser(ctx, target); err != nil {
				return nil, err
			}
			if err := codesRepo.DeleteByUserAndPurpose(ctx, target.ID, "email_verify"); err != nil {
				return nil, err
			}
			return target, nil
		}

		var user *models.User

		if existingByEmail != nil {
			if existingByEmail.IsVerified {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
				return
			}
			if existingByUsername != nil && existingByUsername.ID != existingByEmail.ID {
				c.JSON(http.StatusConflict, gin.H{"error": "Username already in use"})
				return
			}
			user, err = reusePending(existingByEmail)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset existing account"})
				return
			}
		} else {
			if existingByUsername != nil {
				if existingByUsername.IsVerified || strings.ToLower(existingByUsername.Email) != normalizedEmail {
					c.JSON(http.StatusConflict, gin.H{"error": "Username already in use"})
					return
				}
				user, err = reusePending(existingByUsername)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset existing account"})
					return
				}
			} else {
				user = &models.User{
					Username:  req.Username,
					Email:     normalizedEmail,
					Password:  hashed,
					FirstName: req.FirstName,
					LastName:  req.LastName,
				}
				if err := usersRepo.CreateUser(ctx, user); err != nil {
					if isUniqueViolation(err) {
						if existing, lookupErr := usersRepo.GetByEmail(ctx, normalizedEmail); lookupErr == nil && existing != nil && !existing.IsVerified {
							user, err = reusePending(existing)
							if err != nil {
								c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset existing account"})
								return
							}
						} else {
							c.JSON(http.StatusConflict, gin.H{"error": "Email or username already exists"})
							return
						}
					} else {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
						return
					}
				}
			}
		}

		code := generateCode(6)
		v := &models.VerificationCode{
			UserID:    user.ID,
			Purpose:   "email_verify",
			Code:      code,
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}
		if err := codesRepo.Create(ctx, v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verification code"})
			return
		}
		log.Printf("[auth] Sent email verify code to %s (user=%s)", user.Email, user.ID)
		_ = mail.Send(user.Email, "Подтверждение почты", fmt.Sprintf("Ваш код подтверждения: %s", code))

		c.JSON(http.StatusCreated, gin.H{
			"message": "user created; verification code sent to email",
			"user":    user,
		})
	}
}

func Login(usersRepo repository.UserRepository, codesRepo repository.VerificationCodeRepository, mail services.EmailSender, jwtSecret string, email2FAEnabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		req.Email = strings.TrimSpace(req.Email)
		normalizedEmail := strings.ToLower(req.Email)

		user, err := usersRepo.GetByEmail(c.Request.Context(), normalizedEmail)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		req.Password = strings.TrimSpace(req.Password)
		ok, err := auth.VerifyPassword(user.Password, req.Password)
		if err != nil || !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		if !user.IsVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "Email not verified", "action": "verify_email"})
			return
		}

		if email2FAEnabled && user.Is2FAEnabled {
			log.Printf("[auth] 2FA required for login (user=%s, email2FAEnabled=%v, is2FAEnabled=%v); no code sent automatically", user.ID, email2FAEnabled, user.Is2FAEnabled)
			c.JSON(http.StatusOK, gin.H{"status": "2fa_required", "user_id": user.ID})
			return
		}

		token, err := createAccessJWT(jwtSecret, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
			return
		}
		refresh, err := createRefreshJWT(jwtSecret, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign refresh token"})
			return
		}
		setRefreshCookie(c, refresh)
		c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
	}
}

func createAccessJWT(secret string, user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"email":    user.Email,
		"username": user.Username,
		"type":     "access",
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func createRefreshJWT(secret string, user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"type": "refresh",
		"exp":  time.Now().Add(14 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

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

func VerifyEmail(usersRepo repository.UserRepository, codesRepo repository.VerificationCodeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		req.Email = strings.TrimSpace(req.Email)
		req.Code = strings.TrimSpace(req.Code)
		normalizedEmail := strings.ToLower(req.Email)

		user, err := usersRepo.GetByEmail(c.Request.Context(), normalizedEmail)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		v, err := codesRepo.GetValid(c.Request.Context(), user.ID, "email_verify", req.Code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired code"})
			return
		}
		_ = codesRepo.Consume(c.Request.Context(), v.ID)
		user.IsVerified = true
		if err := usersRepo.UpdateUser(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "verified"})
	}
}

func ResendVerificationEmail(usersRepo repository.UserRepository, codesRepo repository.VerificationCodeRepository, mail services.EmailSender) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		req.Email = strings.TrimSpace(req.Email)
		normalizedEmail := strings.ToLower(req.Email)

		user, err := usersRepo.GetByEmail(c.Request.Context(), normalizedEmail)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if user.IsVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Already verified"})
			return
		}
		_ = codesRepo.DeleteByUserAndPurpose(c.Request.Context(), user.ID, "email_verify")
		code := generateCode(6)
		v := &models.VerificationCode{
			UserID:    user.ID,
			Purpose:   "email_verify",
			Code:      code,
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}
		if err := codesRepo.Create(c.Request.Context(), v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verification code"})
			return
		}
		log.Printf("[auth] Resent email verify code to %s (user=%s)", user.Email, user.ID)
		_ = mail.Send(user.Email, "Подтверждение почты", fmt.Sprintf("Ваш код подтверждения: %s", code))
		c.JSON(http.StatusOK, gin.H{"status": "resent"})
	}
}

func VerifyLogin2FA(usersRepo repository.UserRepository, codesRepo repository.VerificationCodeRepository, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		req.Email = strings.TrimSpace(req.Email)
		req.Code = strings.TrimSpace(req.Code)
		normalizedEmail := strings.ToLower(req.Email)

		user, err := usersRepo.GetByEmail(c.Request.Context(), normalizedEmail)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		if !user.Is2FAEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "2fa not enabled"})
			return
		}
		v, err := codesRepo.GetValid(c.Request.Context(), user.ID, "login_2fa", req.Code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired 2fa code"})
			return
		}
		_ = codesRepo.Consume(c.Request.Context(), v.ID)

		token, err := createAccessJWT(jwtSecret, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
			return
		}
		refresh, err := createRefreshJWT(jwtSecret, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign refresh token"})
			return
		}
		setRefreshCookie(c, refresh)
		c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
	}
}

func Request2FACode(usersRepo repository.UserRepository, codesRepo repository.VerificationCodeRepository, mail services.EmailSender, email2FAEnabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if !email2FAEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "2fa by email disabled"})
			return
		}
		req.Email = strings.TrimSpace(req.Email)
		normalizedEmail := strings.ToLower(req.Email)

		user, err := usersRepo.GetByEmail(c.Request.Context(), normalizedEmail)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if !user.IsVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "Email not verified"})
			return
		}
		if !user.Is2FAEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "2fa not enabled"})
			return
		}
		_ = codesRepo.DeleteByUserAndPurpose(c.Request.Context(), user.ID, "login_2fa")
		code := generateCode(6)
		v := &models.VerificationCode{
			UserID:    user.ID,
			Purpose:   "login_2fa",
			Code:      code,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
		if err := codesRepo.Create(c.Request.Context(), v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create 2fa code"})
			return
		}
		log.Printf("[auth] Sent 2FA login code to %s (user=%s, email2FAEnabled=%v, is2FAEnabled=%v)", user.Email, user.ID, email2FAEnabled, user.Is2FAEnabled)
		_ = mail.Send(user.Email, "Код входа", fmt.Sprintf("Your code: %s", code))
		c.JSON(http.StatusOK, gin.H{"status": "sent"})
	}
}
func generateCode(length int) string {
	digits := "0123456789"
	out := make([]byte, length)
	for i := 0; i < length; i++ {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		out[i] = digits[nBig.Int64()]
	}
	return string(out)
}

func isUniqueViolation(err error) bool {
	type sqlStateError interface {
		SQLState() string
	}
	var stateErr sqlStateError
	if errors.As(err, &stateErr) {
		return stateErr.SQLState() == "23505"
	}
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

func isValidEmail(value string) bool {
	if value == "" {
		return false
	}
	_, err := netmail.ParseAddress(value)
	return err == nil
}

func Toggle2FA(usersRepo repository.UserRepository, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Enable bool `json:"enable"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		authz := c.GetHeader("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing bearer token"})
			return
		}
		tokenStr := strings.TrimPrefix(authz, "Bearer ")
		parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !parsed.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok || claims["type"] != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		sub, _ := claims["sub"].(string)
		user, err := usersRepo.GetByID(c.Request.Context(), sub)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		user.Is2FAEnabled = req.Enable
		if err := usersRepo.UpdateUser(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"is_2fa_enabled": user.Is2FAEnabled})
	}
}
