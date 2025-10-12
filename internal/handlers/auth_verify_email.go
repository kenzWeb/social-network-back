package handlers

import (
	"net/http"

	"modern-social-media/internal/services"

	"github.com/gin-gonic/gin"
)

func VerifyEmailWithService(svc services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if err := svc.VerifyEmail(c.Request.Context(), req.Email, req.Code); err != nil {
			if err.Error() == "not_found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			if err.Error() == "invalid_code" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired code"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "verified"})
	}
}
