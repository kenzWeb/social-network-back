package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(rg *gin.RouterGroup, d Deps) {
	rg.GET("/user", handlers.GetAllUsers(d.Models.Users))
	rg.GET("/user/me", middleware.Auth(d.JWTSecret), handlers.GetCurrentUser(d.Models.Users))
	rg.GET("/user/:id", handlers.GetUserById(d.Models.Users))
	rg.GET("/user/by-email/:email", handlers.GetUserByEmail(d.Models.Users))
	rg.POST("/user", middleware.StaticToken(d.AdminToken), handlers.CreateUser(d.Models.Users))
	rg.PUT("/user/:id", handlers.UpdateUser(d.Models.Users))
	rg.DELETE("/user/:id", handlers.DeleteUser(d.Models.Users))
}
