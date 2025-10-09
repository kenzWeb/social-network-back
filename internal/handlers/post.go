package handlers

import (
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetPostsByUser(postRepo repository.PostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, _ := uidAny.(string)

		posts, err := postRepo.GetPostsByUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		println(c)
		c.JSON(http.StatusOK, posts)
	}
}

func CreatePost(postRepo repository.PostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content  string `json:"content"`
			ImageURL string `json:"imageUrl"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})

			return
		}

		if req.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
			return
		}

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidAny.(string)

		post := &models.Post{
			UserID:   userID,
			Content:  req.Content,
			ImageURL: req.ImageURL,
		}

		if err := postRepo.CreatePost(c.Request.Context(), post); err != nil {
			if strings.Contains(err.Error(), "SQLSTATE 23503") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "userId error"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		full, err := postRepo.GetById(c.Request.Context(), post.ID)
		if err != nil {
			c.JSON(http.StatusCreated, post)
			return
		}
		c.JSON(http.StatusCreated, full)
	}
}

func UpdatePost(postRepo repository.PostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req struct {
			Content  string `json:"content"`
			ImageURL string `json:"imageUrl"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})

			return
		}

		if req.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
			return
		}

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidAny.(string)

		post := &models.Post{
			ID:       id,
			UserID:   userID,
			Content:  req.Content,
			ImageURL: req.ImageURL,
		}

		if err := postRepo.UpdatePostByUser(c.Request.Context(), id, userID, post); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "record not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
				return
			}
			if strings.Contains(err.Error(), "SQLSTATE 23503") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "userId error"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		full, err := postRepo.GetById(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusOK, post)
			return
		}
		c.JSON(http.StatusOK, full)
	}
}

func DeletePostByUser(postRepo repository.PostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, _ := uidAny.(string)

		if err := postRepo.DeletePostByUser(c.Request.Context(), id, userID); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "record not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
				return
			}
			if strings.Contains(err.Error(), "SQLSTATE 23503") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "userId error"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
