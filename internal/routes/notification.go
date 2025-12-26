package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(rg *gin.RouterGroup, d Deps) {
	grp := rg.Group("/notifications")
	grp.Use(middleware.Auth(d.JWTSecret))

	grp.GET("", handlers.GetNotifications(d.Models.Notifications))
	grp.GET("/unread", handlers.GetUnreadNotifications(d.Models.Notifications))
	grp.GET("/unread/count", handlers.GetUnreadCount(d.Models.Notifications))
	grp.PATCH("/:id/read", handlers.MarkNotificationAsRead(d.Models.Notifications))
	grp.PATCH("/read-all", handlers.MarkAllNotificationsAsRead(d.Models.Notifications))
	grp.DELETE("/:id", handlers.DeleteNotification(d.Models.Notifications))
}
