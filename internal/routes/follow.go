package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterFollowRoutes(rg *gin.RouterGroup, d Deps) {
	grp := rg.Group("/follow")
	grp.Use(middleware.Auth(d.JWTSecret))
	grp.POST("/:id/toggle", handlers.ToggleFollowWithNotification(d.Models.Follows, d.Models.Notifications))
	grp.POST("/:id", handlers.FollowWithNotification(d.Models.Follows, d.Models.Notifications))
	grp.DELETE("/:id", handlers.UnfollowWithNotification(d.Models.Follows, d.Models.Notifications))
	grp.GET("/:id/status", handlers.IsFollowing(d.Models.Follows))

	rg.GET("/user/:id/followers", handlers.GetFollowers(d.Models.Follows))
	rg.GET("/user/:id/following", handlers.GetFollowing(d.Models.Follows))
	rg.GET("/user/:id/follow-counts", handlers.GetFollowCounts(d.Models.Follows))
}
