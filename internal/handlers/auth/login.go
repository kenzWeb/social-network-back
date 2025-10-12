package auth

import (
	"net/http"

	"modern-social-media/internal/services"

	"github.com/gin-gonic/gin"
)

func LoginWithService(svc services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		user, tokens, err := svc.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			switch err.Error() {
			case "invalid_credentials":
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			case "email_not_verified":
				c.JSON(http.StatusForbidden, gin.H{"error": "Email not verified", "action": "verify_email"})
				return
			case "2fa_required":
				c.JSON(http.StatusOK, gin.H{"status": "2fa_required"})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
				return
			}
		}
		setRefreshCookie(c, tokens.Refresh)
		c.JSON(http.StatusOK, gin.H{"token": tokens.Access, "user": user})
	}
}
