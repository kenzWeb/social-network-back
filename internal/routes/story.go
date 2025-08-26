package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterStoryRoutes(rg *gin.RouterGroup, d Deps) {
	rg.GET("/story/:id", handlers.GetStoryById(d.Models.Stories))
	rg.GET("/user/:userId/stories", handlers.GetStoriesByUserId(d.Models.Stories))

	stories := rg.Group("/story")
	stories.Use(middleware.Auth(d.JWTSecret))
	{
		stories.GET("", handlers.GetStoriesByUser(d.Models.Stories))
		stories.POST("", handlers.CreateStory(d.Models.Stories))
		stories.PUT("/:id", handlers.UpdateStory(d.Models.Stories))
		stories.DELETE("/:id", handlers.DeleteStory(d.Models.Stories))

		stories.POST("/:id/like", handlers.ToggleStoryLike(d.Models.Likes))
	}

	rg.GET("/stories", middleware.Auth(d.JWTSecret), handlers.GetAllStories(d.Models.Stories))
}
