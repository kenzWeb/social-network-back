package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterPostRoutes(rg *gin.RouterGroup, d Deps) {
	rg.GET("/post", middleware.Auth(d.JWTSecret), handlers.GetPostsByUser(d.Models.Posts))

	rg.POST("/post", middleware.Auth(d.JWTSecret), handlers.CreatePost(d.Models.Posts))

	rg.PUT("/post/:id", middleware.Auth(d.JWTSecret), handlers.UpdatePost(d.Models.Posts))

	rg.DELETE("/post/:id", middleware.Auth(d.JWTSecret), handlers.DeletePostByUser(d.Models.Posts))

	rg.POST("/post/:id/like", middleware.Auth(d.JWTSecret), handlers.TogglePostLike(d.Models.Likes))
}
