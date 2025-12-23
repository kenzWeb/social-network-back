package auth

import (
	"errors"
	"net/http"
	"time"

	"modern-social-media/internal/auth"
	"modern-social-media/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func VerifyLogin2FAWithService(svc services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		user, tokens, err := svc.VerifyLogin2FA(c.Request.Context(), req.Email, req.Code)
		if err != nil {
			switch err.Error() {
			case "invalid_credentials":
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			case "2fa_not_enabled":
				c.JSON(http.StatusBadRequest, gin.H{"error": "2fa not enabled"})
			case "invalid_2fa_code":
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired 2fa code"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
			}
			return
		}
		setRefreshCookie(c, tokens.Refresh)
		c.JSON(http.StatusOK, gin.H{"token": tokens.Access, "user": user})
	}
}

func Request2FACodeWithService(svc services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if err := svc.Request2FACode(c.Request.Context(), req.Email); err != nil {
			switch err.Error() {
			case "2fa_disabled", "2fa_disabled_globally":
				c.JSON(http.StatusBadRequest, gin.H{"error": "2fa by email disabled"})
			case "not_found":
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			case "email_not_verified":
				c.JSON(http.StatusForbidden, gin.H{"error": "Email not verified"})
			case "2fa_not_enabled":
				c.JSON(http.StatusBadRequest, gin.H{"error": "2fa not enabled"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create 2fa code"})
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "sent"})
	}
}

func Toggle2FAWithService(svc services.AuthService, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Enable bool `json:"enable"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		sub, err := parseAccessSubject(jwtSecret, c.GetHeader("Authorization"))
		if err != nil {
			msg := err.Error()
			if msg == "missing bearer token" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing bearer token"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			return
		}
		user, err := svc.Users.GetByID(c.Request.Context(), sub)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		user.Is2FAEnabled = req.Enable
		if err := svc.Users.UpdateUser(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"is_2fa_enabled": user.Is2FAEnabled})
	}
}

func Refresh(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("refreshToken")
		if err != nil || cookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
			return
		}

		var claims auth.RefreshClaims
		token, err := jwt.ParseWithClaims(cookie, &claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		if err := claims.Validate(); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
			return
		}

		// Create new access token
		now := time.Now()
		accClaims := auth.AccessClaims{
			UserID: claims.UserID,
			Type:   "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   claims.UserID,
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			},
		}

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accClaims)
		signed, err := accessToken.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": signed})
	}
}
