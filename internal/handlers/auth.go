package handlers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
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

		if u, err := usersRepo.GetByEmail(c.Request.Context(), req.Email); err == nil && u != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check email"})
			return
		}
		if u, err := usersRepo.GetByUsername(c.Request.Context(), req.Username); err == nil && u != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "username already in use"})
			return
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check username"})
			return
		}

		user := &models.User{
			Username:  req.Username,
			Email:     req.Email,
			Password:  "",
			FirstName: req.FirstName,
			LastName:  req.LastName,
		}
		hashed, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = hashed
		if err := usersRepo.CreateUser(c.Request.Context(), user); err != nil {
			if strings.Contains(err.Error(), "SQLSTATE 23505") {
				c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		code := generateCode(6)
		v := &models.VerificationCode{
			UserID:    user.ID,
			Purpose:   "email_verify",
			Code:      code,
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}
		if err := codesRepo.Create(c.Request.Context(), v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create verification code"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		user, err := usersRepo.GetByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		ok, err := auth.VerifyPassword(user.Password, req.Password)
		if err != nil || !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if !user.IsVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "email not verified", "action": "verify_email"})
			return
		}

		if email2FAEnabled && user.Is2FAEnabled {
			log.Printf("[auth] 2FA required for login (user=%s, email2FAEnabled=%v, is2FAEnabled=%v); no code sent automatically", user.ID, email2FAEnabled, user.Is2FAEnabled)
			c.JSON(http.StatusOK, gin.H{"status": "2fa_required", "user_id": user.ID})
			return
		}

		token, err := createAccessJWT(jwtSecret, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}
		refresh, err := createRefreshJWT(jwtSecret, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign refresh token"})
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
			return
		}
		parsed, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !parsed.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok || claims["type"] != "refresh" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		user, err := usersRepo.GetByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		v, err := codesRepo.GetValid(c.Request.Context(), user.ID, "email_verify", req.Code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired code"})
			return
		}
		_ = codesRepo.Consume(c.Request.Context(), v.ID)
		user.IsVerified = true
		if err := usersRepo.UpdateUser(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		user, err := usersRepo.GetByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if user.IsVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "already verified"})
			return
		}
		code := generateCode(6)
		v := &models.VerificationCode{
			UserID:    user.ID,
			Purpose:   "email_verify",
			Code:      code,
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}
		if err := codesRepo.Create(c.Request.Context(), v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create verification code"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		user, err := usersRepo.GetByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if !user.Is2FAEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "2fa not enabled"})
			return
		}
		v, err := codesRepo.GetValid(c.Request.Context(), user.ID, "login_2fa", req.Code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired 2fa code"})
			return
		}
		_ = codesRepo.Consume(c.Request.Context(), v.ID)

		token, err := createAccessJWT(jwtSecret, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}
		refresh, err := createRefreshJWT(jwtSecret, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign refresh token"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if !email2FAEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "2fa by email disabled"})
			return
		}
		user, err := usersRepo.GetByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if !user.IsVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
			return
		}
		if !user.Is2FAEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "2fa not enabled"})
			return
		}
		code := generateCode(6)
		v := &models.VerificationCode{
			UserID:    user.ID,
			Purpose:   "login_2fa",
			Code:      code,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
		if err := codesRepo.Create(c.Request.Context(), v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create 2fa code"})
			return
		}
		log.Printf("[auth] Sent 2FA login code to %s (user=%s, email2FAEnabled=%v, is2FAEnabled=%v)", user.Email, user.ID, email2FAEnabled, user.Is2FAEnabled)
		_ = mail.Send(user.Email, "Код входа", fmt.Sprintf("Ваш код входа: %s", code))
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

func Toggle2FA(usersRepo repository.UserRepository, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Enable bool `json:"enable"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		authz := c.GetHeader("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok || claims["type"] != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"is_2fa_enabled": user.Is2FAEnabled})
	}
}
