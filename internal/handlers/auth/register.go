package auth

import (
	"net/http"

	"modern-social-media/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterWithService(svc services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username  string `json:"username"`
			Email     string `json:"email"`
			Password  string `json:"password"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		user, err := svc.Register(c.Request.Context(), services.RegisterInput{
			Username:  req.Username,
			Email:     req.Email,
			Password:  req.Password,
			FirstName: req.FirstName,
			LastName:  req.LastName,
		})
		if err != nil {
			if err.Error() == "email_in_use" {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
				return
			}
			if err.Error() == "username_in_use" {
				c.JSON(http.StatusConflict, gin.H{"error": "Username already in use"})
				return
			}
			if err.Error() == "invalid_email" || err.Error() == "missing_required_fields" || err.Error() == "weak_password" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "User created; verification code sent to your Email", "user": user})
	}
}
