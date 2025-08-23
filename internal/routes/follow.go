package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterFollowRoutes(rg *gin.RouterGroup, d Deps) {
	grp := rg.Group("/follow")
	grp.Use(middleware.Auth(d.JWTSecret))
	grp.POST("/:id/toggle", handlers.ToggleFollow(d.Models.Follows))
	grp.POST("/:id", handlers.Follow(d.Models.Follows))
	grp.DELETE("/:id", handlers.Unfollow(d.Models.Follows))
	grp.GET("/:id/status", handlers.IsFollowing(d.Models.Follows))

	rg.GET("/user/:id/followers", handlers.GetFollowers(d.Models.Follows))
	rg.GET("/user/:id/following", handlers.GetFollowing(d.Models.Follows))
	rg.GET("/user/:id/follow-counts", handlers.GetFollowCounts(d.Models.Follows))
}
