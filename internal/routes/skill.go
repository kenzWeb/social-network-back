package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterSkillRoutes(rg *gin.RouterGroup, d Deps) {
	rg.GET("/skill", middleware.Auth(d.JWTSecret), handlers.GetAllSkills(d.Models.Skills))

	rg.POST("/skill", middleware.Auth(d.JWTSecret), handlers.AddSkill(d.Models.Skills))
}
