package handlers

import (
	"errors"
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetCommentsByPost(commentRepo repository.CommentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID := c.Param("id")
		if postID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "post id is required"})
			return
		}
		comments, err := commentRepo.GetCommentsByPost(c.Request.Context(), postID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve comments"})
			return
		}
		c.JSON(http.StatusOK, comments)
	}
}

func GetCommentById(commentRepo repository.CommentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID := c.Param("id")
		if commentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "comment id is required"})
			return
		}
		comment, err := commentRepo.GetCommentById(c.Request.Context(), commentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve comment"})
			return
		}
		c.JSON(http.StatusOK, comment)
	}
}

func GetCommentsByUser(commentRepo repository.CommentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidAny.(string)

		comments, err := commentRepo.GetCommentsByUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve comments"})
			return
		}
		c.JSON(http.StatusOK, comments)
	}
}

func CreateComment(commentRepo repository.CommentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Message string `json:"message"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidAny.(string)

		postID := c.Param("id")
		if postID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "post id is required"})
			return
		}
		comment := &models.Comment{
			UserID:  userID,
			PostID:  postID,
			Message: req.Message,
		}

		if err := commentRepo.CreateComment(c.Request.Context(), comment); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
			return
		}

		c.JSON(http.StatusCreated, comment)
	}
}
