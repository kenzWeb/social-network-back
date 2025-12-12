package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(rg *gin.RouterGroup, d Deps) {
	rg.GET("/user/me", middleware.Auth(d.JWTSecret), handlers.GetCurrentUser(d.Models.Users))
	rg.GET("/user/me/followers", middleware.Auth(d.JWTSecret), handlers.GetMyFollowers(d.Models.Follows))
	rg.GET("/user/me/following", middleware.Auth(d.JWTSecret), handlers.GetMyFollowing(d.Models.Follows))
	rg.GET("/user/by-email/:email", handlers.GetUserByEmail(d.Models.Users))

	protected := rg.Group("", middleware.StaticToken(d.AdminToken))
	{
		protected.GET("/user", handlers.GetAllUsers(d.Models.Users))
		protected.GET("/user/:id", handlers.GetUserById(d.Models.Users))
		protected.POST("/user", handlers.CreateUser(d.Models.Users))
		protected.PUT("/user/:id", handlers.UpdateUser(d.Models.Users))
		protected.DELETE("/user/:id", handlers.DeleteUser(d.Models.Users))
	}
}
