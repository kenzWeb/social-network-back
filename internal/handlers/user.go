package handlers

import (
	"net/http"
	"reflect"

	"modern-social-media/internal/auth"
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
	"modern-social-media/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetAllUsers(usersRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := usersRepo.GetAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}
		c.JSON(http.StatusOK, users)
	}
}

func GetCurrentUser(usersRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidAny.(string)

		user, err := usersRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User Not Found"})
			return
		}
		c.JSON(http.StatusOK, user)
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		hashed, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		user := &models.User{
			Username:  req.Username,
			Email:     req.Email,
			Password:  hashed,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			AvatarURL: utils.GetRandomDefaultAvatar(),
		}

		if err := usersRepo.CreateUser(c.Request.Context(), user); err != nil {
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
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var req struct {
			Username     *string `json:"username"`
			Email        *string `json:"email"`
			Password     *string `json:"password"`
			FirstName    *string `json:"first_name"`
			LastName     *string `json:"last_name"`
			Bio          *string `json:"bio"`
			AvatarURL    *string `json:"avatar_url"`
			IsActive     *bool   `json:"is_active"`
			IsVerified   *bool   `json:"is_verified"`
			Is2FAEnabled *bool   `json:"is_2fa_enabled"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if req.Password != nil {
			hashed, err := auth.HashPassword(*req.Password)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
				return
			}
			existing.Password = hashed
			req.Password = nil
		}

		applyUserPatch(existing, req)

		if err := usersRepo.UpdateUser(c.Request.Context(), existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, existing)
	}
}

func DeleteUser(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := userRepo.Delete(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func applyUserPatch(existing *models.User, patch interface{}) {
	rv := reflect.ValueOf(patch)
	if rv.Kind() != reflect.Struct {
		return
	}
	ev := reflect.ValueOf(existing).Elem()
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		if f.Kind() != reflect.Ptr || f.IsNil() {
			continue
		}
		fieldName := rt.Field(i).Name
		target := ev.FieldByName(fieldName)
		if !target.IsValid() || !target.CanSet() {
			continue
		}
		target.Set(f.Elem())
	}
}
