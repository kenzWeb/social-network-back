package handlers

import (
	"errors"
	"modern-social-media/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type likeResponse struct {
	Liked      bool  `json:"liked"`
	LikesCount int64 `json:"likes_count"`
}

func TogglePostLike(likeRepo repository.LikeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidAny.(string)
		postID := c.Param("id")
		if postID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Post id is required"})
			return
		}
		liked, err := likeRepo.TogglePostLike(c.Request.Context(), userID, postID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle like"})
			return
		}
		cnt, err := likeRepo.CountPostLikes(c.Request.Context(), postID)
		if err != nil {
			c.JSON(http.StatusOK, likeResponse{Liked: liked, LikesCount: 0})
			return
		}
		c.JSON(http.StatusOK, likeResponse{Liked: liked, LikesCount: cnt})
	}
}

func ToggleStoryLike(likeRepo repository.LikeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidAny.(string)
		storyID := c.Param("id")
		if storyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Story id is required"})
			return
		}
		liked, err := likeRepo.ToggleStoryLike(c.Request.Context(), userID, storyID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle like"})
			return
		}
		cnt, err := likeRepo.CountStoryLikes(c.Request.Context(), storyID)
		if err != nil {
			c.JSON(http.StatusOK, likeResponse{Liked: liked, LikesCount: 0})
			return
		}
		c.JSON(http.StatusOK, likeResponse{Liked: liked, LikesCount: cnt})
	}
}
