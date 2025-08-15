package main

import (
	"modern-social-media/internal/handlers"
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
	v1.POST("/auth/register", handlers.Register(app.models.Users, app.jwtSecret))
	v1.POST("/auth/login", handlers.Login(app.models.Users, app.jwtSecret))
	v1.POST("/auth/refresh", handlers.Refresh(app.jwtSecret))

	return g
}
