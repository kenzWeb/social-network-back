package handlers

import (
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
	"net/http"
	"reflect"

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

		applyUserPatch(existing, req)

		if err := usersRepo.Update(c.Request.Context(), existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, existing)
	}
}

// applyUserPatch применяет непустые (не nil) поля структуры patch к existing через reflection.
// Предполагается, что имена экспортируемых полей patch совпадают с именами полей модели User.
func applyUserPatch(existing *models.User, patch interface{}) {
	rv := reflect.ValueOf(patch)
	// работаем только со структурой (значением)
	if rv.Kind() != reflect.Struct {
		return
	}
	ev := reflect.ValueOf(existing).Elem()
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		if f.Kind() != reflect.Ptr || f.IsNil() { // нужны только указатели и не nil
			continue
		}
		fieldName := rt.Field(i).Name
		target := ev.FieldByName(fieldName)
		if !target.IsValid() || !target.CanSet() {
			continue
		}
		// Устанавливаем разыменованное значение
		target.Set(f.Elem())
	}
}
