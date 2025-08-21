package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterPostRoutes(rg *gin.RouterGroup, d Deps) {
	rg.POST("/post", middleware.Auth(d.JWTSecret), handlers.CreatePost(d.Models.Posts))
}
