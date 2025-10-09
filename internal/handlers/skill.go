package handlers

import (
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllSkills(skillRepo repository.SkillRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)
		skills, err := skillRepo.GetAllSkills(c.Request.Context(), userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, skills)
	}
}

func AddSkill(skillRepo repository.SkillRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Skill name is required"})
			return
		}

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidAny.(string)

		skill := &models.Skill{
			UserID: userID,
			Name:   req.Name,
		}

		if err := skillRepo.AddSkill(c.Request.Context(), skill); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add skill"})
			return
		}
	}
}
