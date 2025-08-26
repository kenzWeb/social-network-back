package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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
		file, err := c.FormFile("media")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "media file is required"})
			return
		}

		contentType := file.Header.Get("Content-Type")
		var mediaType string
		var maxSize int64

		if strings.HasPrefix(contentType, "image/") {
			mediaType = "image"
			maxSize = 100 * 1024 * 1024
		} else if strings.HasPrefix(contentType, "video/") {
			mediaType = "video"
			maxSize = 100 * 1024 * 1024
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only image and video files are allowed"})
			return
		}

		if file.Size > maxSize {
			maxSizeMB := maxSize / (1024 * 1024)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file size must be less than %dMB", maxSizeMB)})
			return
		}

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)
		filename := fmt.Sprintf("%s_%d_%s", userID, time.Now().Unix(), file.Filename)
		uploadPath := fmt.Sprintf("uploads/stories/%s", filename)

		if err := os.MkdirAll("uploads/stories", 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
			return
		}

		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
			return
		}

		story := &models.Story{
			UserID:    userID,
			MediaURL:  "/" + uploadPath,
			MediaType: mediaType,
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

		file, err := c.FormFile("media")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "media file is required"})
			return
		}

		contentType := file.Header.Get("Content-Type")
		var mediaType string
		var maxSize int64

		if strings.HasPrefix(contentType, "image/") {
			mediaType = "image"
			maxSize = 100 * 1024 * 1024
		} else if strings.HasPrefix(contentType, "video/") {
			mediaType = "video"
			maxSize = 100 * 1024 * 1024
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only image and video files are allowed"})
			return
		}

		if file.Size > maxSize {
			maxSizeMB := maxSize / (1024 * 1024)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file size must be less than %dMB", maxSizeMB)})
			return
		}

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		filename := fmt.Sprintf("%s_%d_%s", userID, time.Now().Unix(), file.Filename)
		uploadPath := fmt.Sprintf("uploads/stories/%s", filename)

		if err := os.MkdirAll("uploads/stories", 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
			return
		}

		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
			return
		}

		story := &models.Story{
			MediaURL:  "/" + uploadPath,
			MediaType: mediaType,
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
