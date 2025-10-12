package handlers

import (
	"net/http"

	"modern-social-media/internal/services"

	"github.com/gin-gonic/gin"
)

func ResendVerificationEmailWithService(svc services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if err := svc.ResendVerificationEmail(c.Request.Context(), req.Email); err != nil {
			switch err.Error() {
			case "not_found":
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			case "already_verified":
				c.JSON(http.StatusBadRequest, gin.H{"error": "Already verified"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verification code"})
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "resent"})
	}
}
