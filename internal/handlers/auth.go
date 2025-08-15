package handlers

import (
	"errors"
	"net/http"
	"time"

	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"modern-social-media/internal/auth"
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
)

func Register(usersRepo repository.UserRepository, jwtSecret string) gin.HandlerFunc {
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
		if err := usersRepo.Create(c.Request.Context(), user); err != nil {
			if strings.Contains(err.Error(), "SQLSTATE 23505") {
				c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

		c.JSON(http.StatusCreated, gin.H{
			"token": token,
			"user":  user,
		})
	}
}

func Login(usersRepo repository.UserRepository, jwtSecret string) gin.HandlerFunc {
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

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user":  user,
		})
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
		true, // secure: true — на проде за HTTPS; можно сделать зависимым от окружения
		true, // httpOnly
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
