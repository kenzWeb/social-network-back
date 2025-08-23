package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterStoryRoutes(rg *gin.RouterGroup, d Deps) {
	rg.POST("/story/:id/like", middleware.Auth(d.JWTSecret), handlers.ToggleStoryLike(d.Models.Likes))
}
