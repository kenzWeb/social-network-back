package main

import (
	introutes "modern-social-media/internal/routes"
	"net/http"

	"modern-social-media/internal/handlers"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	g := gin.New()
	g.Use(gin.Logger(), gin.Recovery())

	g.Static("/uploads", "./uploads")

	v1 := g.Group("/api/v1")

	deps := introutes.Deps{
		Models:          app.models,
		Mailer:          app.mailer,
		JWTSecret:       app.jwtSecret,
		AdminToken:      app.adminToken,
		Email2FAEnabled: app.email2FAEnabled,
	}
	introutes.RegisterUserRoutes(v1, deps)
	introutes.RegisterAuthRoutes(v1, deps)
	introutes.RegisterPostRoutes(v1, deps)
	introutes.RegisterCommentsRoutes(v1, deps)
	introutes.RegisterStoryRoutes(v1, deps)
	introutes.RegisterFollowRoutes(v1, deps)
	introutes.RegisterSkillRoutes(v1, deps)

	hub := handlers.NewHub()
	introutes.RegisterChatRoutes(v1, deps, hub)

	return g
}
