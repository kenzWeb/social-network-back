package main

import (
	introutes "modern-social-media/internal/routes"
	"net/http"
	"time"

	"modern-social-media/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (app *application) routes() http.Handler {
	g := gin.New()
	g.Use(gin.Logger(), gin.Recovery())

	g.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	g.Static("/uploads", "./uploads")

	g.GET("/openapi.json", func(c *gin.Context) {
		c.File("./openapi.json")
	})

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/openapi.json")))

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
