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

// @name StoryResponse
type StoryResponse struct {
	ID        string    `json:"id"`
	MediaURL  string    `json:"media_url"`
	MediaType string    `json:"media_type"`
	CreatedAt time.Time `json:"created_at"`
}

// @name UserStoriesResponse
type UserStoriesResponse struct {
	ID        string          `json:"id"`
	Username  string          `json:"username"`
	AvatarURL string          `json:"avatar_url"`
	Stories   []StoryResponse `json:"stories"`
}

// @Summary Get stories feed
// @Description Get stories from followed users
// @Tags stories
// @Produce json
// @Success 200 {array} UserStoriesResponse
// @Router /story/following [get]
func GetAllStories(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		users, err := storyRepo.GetFollowedUsersWithStories(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []UserStoriesResponse
		for _, user := range users {
			var storyResponses []StoryResponse
			for _, s := range user.Stories {
				storyResponses = append(storyResponses, StoryResponse{
					ID:        s.ID,
					MediaURL:  utils.GetFullURL(s.MediaURL),
					MediaType: s.MediaType,
					CreatedAt: s.CreatedAt,
				})
			}
			response = append(response, UserStoriesResponse{
				ID:        user.ID,
				Username:  user.Username,
				AvatarURL: utils.GetFullURL(user.AvatarURL),
				Stories:   storyResponses,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

func GetStoryById(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		story, err := storyRepo.GetById(c.Request.Context(), id)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "record not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, StoryResponse{
			ID:        story.ID,
			MediaURL:  utils.GetFullURL(story.MediaURL),
			MediaType: story.MediaType,
			CreatedAt: story.CreatedAt,
		})
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

		var response []StoryResponse
		for _, s := range stories {
			response = append(response, StoryResponse{
				ID:        s.ID,
				MediaURL:  utils.GetFullURL(s.MediaURL),
				MediaType: s.MediaType,
				CreatedAt: s.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, response)
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

		var response []StoryResponse
		for _, s := range stories {
			response = append(response, StoryResponse{
				ID:        s.ID,
				MediaURL:  utils.GetFullURL(s.MediaURL),
				MediaType: s.MediaType,
				CreatedAt: s.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

// @Summary Create a new story
// @Description Upload a new story (image or video)
// @Tags stories
// @Accept multipart/form-data
// @Produce json
// @Param media formData file true "Story media file"
// @Success 201 {object} StoryResponse
// @Router /story [post]
func CreateStory(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("media")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Media file is required"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only image and video files are allowed"})
			return
		}

		if file.Size > maxSize {
			maxSizeMB := maxSize / (1024 * 1024)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File size must be less than %dMB", maxSizeMB)})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
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
			// Fallback if fetch fails
			c.JSON(http.StatusCreated, StoryResponse{
				ID:        story.ID,
				MediaURL:  utils.GetFullURL(story.MediaURL),
				MediaType: story.MediaType,
				CreatedAt: story.CreatedAt,
			})
			return
		}

		c.JSON(http.StatusCreated, StoryResponse{
			ID:        fullStory.ID,
			MediaURL:  utils.GetFullURL(fullStory.MediaURL),
			MediaType: fullStory.MediaType,
			CreatedAt: fullStory.CreatedAt,
		})
	}
}

func UpdateStory(storyRepo repository.StoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		file, err := c.FormFile("media")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Media file is required"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only image and video files are allowed"})
			return
		}

		if file.Size > maxSize {
			maxSizeMB := maxSize / (1024 * 1024)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File size must be less than %dMB", maxSizeMB)})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		story := &models.Story{
			MediaURL:  "/" + uploadPath,
			MediaType: mediaType,
		}

		if err := storyRepo.UpdateStoryByUser(c.Request.Context(), id, userID, story); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "record not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fullStory, err := storyRepo.GetById(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusOK, StoryResponse{
				ID:        story.ID,
				MediaURL:  utils.GetFullURL(story.MediaURL),
				MediaType: story.MediaType,
				CreatedAt: story.CreatedAt,
			})
			return
		}

		c.JSON(http.StatusOK, StoryResponse{
			ID:        fullStory.ID,
			MediaURL:  utils.GetFullURL(fullStory.MediaURL),
			MediaType: fullStory.MediaType,
			CreatedAt: fullStory.CreatedAt,
		})
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
				c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
