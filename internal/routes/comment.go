package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterCommentsRoutes(rg *gin.RouterGroup, d Deps) {
	rg.GET("/comment/post/:id", handlers.GetCommentsByPost(d.Models.Comments))

	rg.GET("/comment/user/", middleware.Auth(d.JWTSecret), handlers.GetCommentsByUser(d.Models.Comments))

	rg.GET("/comment/:id", handlers.GetCommentById(d.Models.Comments))

	rg.POST("/comment/post/:id", middleware.Auth(d.JWTSecret), handlers.CreateComment(d.Models.Comments))
}
