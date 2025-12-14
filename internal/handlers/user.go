package handlers

import (
	"net/http"
	"reflect"
	"time"

	"modern-social-media/internal/auth"
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
	"modern-social-media/internal/utils"

	"github.com/gin-gonic/gin"
)

// @name UserDTO
type UserDTO struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Bio            string    `json:"bio"`
	AvatarURL      string    `json:"avatar_url"`
	IsVerified     bool      `json:"is_verified"`
	IsActive       bool      `json:"is_active"`
	Is2FAEnabled   bool      `json:"is_2fa_enabled"`
	FollowersCount int64     `json:"followers_count"`
	FollowingCount int64     `json:"following_count"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

// @name CreateUserRequest
type CreateUserRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// @name UpdateUserRequest
type UpdateUserRequest struct {
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


// @Summary Get all users
// @Description Get a list of all users
// @Tags users
// @Produce json
// @Success 200 {array} UserDTO
// @Router /users [get]
func GetAllUsers(usersRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := usersRepo.GetAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}

		var dtos []UserDTO
		for _, u := range users {
			followersCount, _ := usersRepo.GetFollowersCount(c.Request.Context(), u.ID)
			followingCount, _ := usersRepo.GetFollowingCount(c.Request.Context(), u.ID)
			dtos = append(dtos, UserDTO{
				ID:             u.ID,
				Username:       u.Username,
				Email:          u.Email,
				FirstName:      u.FirstName,
				LastName:       u.LastName,
				Bio:            u.Bio,
				AvatarURL:      utils.GetFullURL(u.AvatarURL),
				IsVerified:     u.IsVerified,
				IsActive:       u.IsActive,
				Is2FAEnabled:   u.Is2FAEnabled,
				FollowersCount: followersCount,
				FollowingCount: followingCount,
				CreatedAt:      u.CreatedAt.Format(time.RFC3339),
				UpdatedAt:      u.UpdatedAt.Format(time.RFC3339),
			})
		}

		c.JSON(http.StatusOK, dtos)
	}
}

// @Summary Get current user
// @Description Get the currently authenticated user
// @Tags users
// @Produce json
// @Success 200 {object} UserDTO
// @Router /auth/me [get]
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

		followersCount, _ := usersRepo.GetFollowersCount(c.Request.Context(), userID)
		followingCount, _ := usersRepo.GetFollowingCount(c.Request.Context(), userID)

		dto := UserDTO{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Bio:            user.Bio,
			AvatarURL:      utils.GetFullURL(user.AvatarURL),
			IsVerified:     user.IsVerified,
			IsActive:       user.IsActive,
			Is2FAEnabled:   user.Is2FAEnabled,
			FollowersCount: followersCount,
			FollowingCount: followingCount,
			CreatedAt:      user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      user.UpdatedAt.Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, dto)
	}
}

// @Summary Get user by ID
// @Description Get user details by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserDTO
// @Router /users/{id} [get]
func GetUserById(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		user, err := userRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User Not Found"})
			return
		}

		followersCount, _ := userRepo.GetFollowersCount(c.Request.Context(), id)
		followingCount, _ := userRepo.GetFollowingCount(c.Request.Context(), id)
		
		dto := UserDTO{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Bio:            user.Bio,
			AvatarURL:      utils.GetFullURL(user.AvatarURL),
			IsVerified:     user.IsVerified,
			IsActive:       user.IsActive,
			Is2FAEnabled:   user.Is2FAEnabled,
			FollowersCount: followersCount,
			FollowingCount: followingCount,
			CreatedAt:      user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      user.UpdatedAt.Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, dto)
	}
}

// @Summary Get user by email
// @Description Get user details by email
// @Tags users
// @Produce json
// @Param email path string true "User Email"
// @Success 200 {object} UserDTO
// @Router /users/email/{email} [get]
func GetUserByEmail(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Param("email")
		user, err := userRepo.GetByEmail(c.Request.Context(), email)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User Not Found"})
			return
		}
		// Since GetByEmail also returns *models.User, we should map it to DTO ideally,
		// but let's assume for now keeping it consistent.
		// NOTE: User model has passwords and stuff, DTO is safer.
		followersCount, _ := userRepo.GetFollowersCount(c.Request.Context(), user.ID)
		followingCount, _ := userRepo.GetFollowingCount(c.Request.Context(), user.ID)
		
		dto := UserDTO{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Bio:            user.Bio,
			AvatarURL:      utils.GetFullURL(user.AvatarURL),
			IsVerified:     user.IsVerified,
			IsActive:       user.IsActive,
			Is2FAEnabled:   user.Is2FAEnabled,
			FollowersCount: followersCount,
			FollowingCount: followingCount,
			CreatedAt:      user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      user.UpdatedAt.Format(time.RFC3339),
		}
		c.JSON(http.StatusOK, dto)
	}
}

// @Summary Create user
// @Description Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User registration details"
// @Success 201 {object} UserDTO
// @Router /auth/register [post]
func CreateUser(usersRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
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
		
		dto := UserDTO{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Bio:            user.Bio,
			AvatarURL:      utils.GetFullURL(user.AvatarURL),
			IsVerified:     user.IsVerified,
			IsActive:       user.IsActive,
			Is2FAEnabled:   user.Is2FAEnabled,
			CreatedAt:      user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      user.UpdatedAt.Format(time.RFC3339),
		}

		c.JSON(http.StatusCreated, dto)
	}
}

// @Summary Update user
// @Description Update user details
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "User update details"
// @Success 200 {object} UserDTO
// @Router /users/{id} [put]
func UpdateUser(usersRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		existing, err := usersRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var req UpdateUserRequest
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
		
		followersCount, _ := usersRepo.GetFollowersCount(c.Request.Context(), existing.ID)
		followingCount, _ := usersRepo.GetFollowingCount(c.Request.Context(), existing.ID)

		dto := UserDTO{
			ID:             existing.ID,
			Username:       existing.Username,
			Email:          existing.Email,
			FirstName:      existing.FirstName,
			LastName:       existing.LastName,
			Bio:            existing.Bio,
			AvatarURL:      utils.GetFullURL(existing.AvatarURL),
			IsVerified:     existing.IsVerified,
			IsActive:       existing.IsActive,
			Is2FAEnabled:   existing.Is2FAEnabled,
			FollowersCount: followersCount,
			FollowingCount: followingCount,
			CreatedAt:      existing.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      existing.UpdatedAt.Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, dto)
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
