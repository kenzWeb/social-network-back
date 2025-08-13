package handlers

import (
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(usersRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := usersRepo.GetAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
			return
		}
		c.JSON(http.StatusOK, users)
	}
}

func GetUserById(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		user, err := userRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User Not Found"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func GetUserByEmail(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Param("email")
		user, err := userRepo.GetByEmail(c.Request.Context(), email)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User Not Found"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func CreateUser(usersRepo repository.UserRepository) gin.HandlerFunc {
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

		user := &models.User{
			Username:  req.Username,
			Email:     req.Email,
			Password:  req.Password,
			FirstName: req.FirstName,
			LastName:  req.LastName,
		}

		if err := usersRepo.Create(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, user)
	}
}

func UpdateUser(usersRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		existing, err := usersRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		var req struct {
			Username  *string `json:"username"`
			Email     *string `json:"email"`
			Password  *string `json:"password"`
			FirstName *string `json:"first_name"`
			LastName  *string `json:"last_name"`
			Bio       *string `json:"bio"`
			AvatarURL *string `json:"avatar_url"`
			IsActive  *bool   `json:"is_active"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if req.Username != nil {
			existing.Username = *req.Username
		}
		if req.Email != nil {
			existing.Email = *req.Email
		}
		if req.Password != nil {
			existing.Password = *req.Password 
		}
		if req.FirstName != nil {
			existing.FirstName = *req.FirstName
		}
		if req.LastName != nil {
			existing.LastName = *req.LastName
		}
		if req.Bio != nil {
			existing.Bio = *req.Bio
		}
		if req.AvatarURL != nil {
			existing.AvatarURL = *req.AvatarURL
		}
		if req.IsActive != nil {
			existing.IsActive = *req.IsActive
		}

		if err := usersRepo.Update(c.Request.Context(), existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, existing)
	}
}
