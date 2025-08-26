package handlers

import (
	"net/http"
	"strings"

	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"

	"github.com/gin-gonic/gin"
)

func GetAllStories(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		stories, err := storyRepo.GetStoriesFromFollowing(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stories)
	}
}

func GetStoryById(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		story, err := storyRepo.GetById(c.Request.Context(), id)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "record not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, story)
	}
}

func GetStoriesByUser(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		stories, err := storyRepo.GetStoriesByUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stories)
	}
}

func GetStoriesByUserId(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		stories, err := storyRepo.GetRecentStoriesByUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stories)
	}
}

func CreateStory(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ContentURL string `json:"content_url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		story := &models.Story{
			UserID:     userID,
			ContentURL: req.ContentURL,
		}

		if err := storyRepo.CreateStory(c.Request.Context(), story); err != nil {
			if strings.Contains(err.Error(), "SQLSTATE 23503") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fullStory, err := storyRepo.GetById(c.Request.Context(), story.ID)
		if err != nil {
			c.JSON(http.StatusCreated, story)
			return
		}

		c.JSON(http.StatusCreated, fullStory)
	}
}

func UpdateStory(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req struct {
			ContentURL string `json:"content_url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		story := &models.Story{
			ContentURL: req.ContentURL,
		}

		if err := storyRepo.UpdateStoryByUser(c.Request.Context(), id, userID, story); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "record not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fullStory, err := storyRepo.GetById(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusOK, story)
			return
		}

		c.JSON(http.StatusOK, fullStory)
	}
}

func DeleteStory(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		if err := storyRepo.DeleteStoryByUser(c.Request.Context(), id, userID); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "record not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
