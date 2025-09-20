package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterChatRoutes(rg *gin.RouterGroup, d Deps, hub *handlers.Hub) {
	rg.GET("/ws", handlers.ChatWSHandler(handlers.ChatWSDeps{Models: d.Models, JWTSecret: d.JWTSecret, Hub: hub}))

	chat := rg.Group("/chat")
	chat.Use(middleware.Auth(d.JWTSecret))
	{
		chat.GET("/conversations", handlers.ListConversations(d.Models))
		chat.GET("/conversations/:id/messages", handlers.ListMessages(d.Models))
		chat.POST("/direct/:user_id/send", handlers.SendDirectMessage(d.Models, hub))
		chat.POST("/conversations/:id/read", handlers.MarkRead(d.Models))
		chat.GET("/presence/:user_id", handlers.GetPresence(hub))
	}
}
