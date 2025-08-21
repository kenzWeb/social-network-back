package main

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	g := gin.New()
	g.Use(gin.Logger(), gin.Recovery())

	v1 := g.Group("/api/v1")

	// user
	v1.GET("/user", handlers.GetAllUsers(app.models.Users))
	v1.GET("/user/:id", handlers.GetUserById(app.models.Users))
	v1.GET("/user/by-email/:email", handlers.GetUserByEmail(app.models.Users))
	v1.POST("/user", handlers.CreateUser(app.models.Users))
	v1.PUT("/user/:id", handlers.UpdateUser(app.models.Users))
	v1.DELETE("/user/:id", handlers.DeleteUser(app.models.Users))

	// auth
	v1.POST("/auth/register", handlers.Register(app.models.Users, app.models.VerificationCodes, app.mailer, app.jwtSecret))
	v1.POST("/auth/login", handlers.Login(app.models.Users, app.models.VerificationCodes, app.mailer, app.jwtSecret, app.email2FAEnabled))
	v1.POST("/auth/refresh", handlers.Refresh(app.jwtSecret))
	v1.POST("/auth/verify-email", handlers.VerifyEmail(app.models.Users, app.models.VerificationCodes))
	v1.POST("/auth/resend-verify-email", handlers.ResendVerificationEmail(app.models.Users, app.models.VerificationCodes, app.mailer))
	v1.POST("/auth/2fa/verify", handlers.VerifyLogin2FA(app.models.Users, app.models.VerificationCodes, app.jwtSecret))
	v1.POST("/auth/2fa/request", handlers.Request2FACode(app.models.Users, app.models.VerificationCodes, app.mailer, app.email2FAEnabled))
	v1.POST("/auth/toggle-2fa", handlers.Toggle2FA(app.models.Users, app.jwtSecret))

	// post
	v1.POST("/post", middleware.Auth(app.jwtSecret), handlers.CreatePost(app.models.Posts))

	return g
}
