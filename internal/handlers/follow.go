package handlers

import (
	"net/http"
	"strconv"

	"modern-social-media/internal/repository"

	"github.com/gin-gonic/gin"
)

type followStatusResponse struct {
	Following      bool  `json:"following"`
	Followers      int64 `json:"followers,omitempty"`
	FollowingCount int64 `json:"following_count,omitempty"`
}

func ToggleFollow(repo repository.FollowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		followerID := uidAny.(string)
		targetID := c.Param("id")
		following, err := repo.ToggleFollow(c.Request.Context(), followerID, targetID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		followers, _ := repo.CountFollowers(c.Request.Context(), targetID)
		c.JSON(http.StatusOK, followStatusResponse{Following: following, Followers: followers})
	}
}

func Follow(repo repository.FollowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		followerID := uidAny.(string)
		targetID := c.Param("id")
		_, err := repo.Follow(c.Request.Context(), followerID, targetID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		followers, _ := repo.CountFollowers(c.Request.Context(), targetID)
		c.JSON(http.StatusOK, followStatusResponse{Following: true, Followers: followers})
	}
}

func Unfollow(repo repository.FollowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		followerID := uidAny.(string)
		targetID := c.Param("id")
		_, err := repo.Unfollow(c.Request.Context(), followerID, targetID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		followers, _ := repo.CountFollowers(c.Request.Context(), targetID)
		c.JSON(http.StatusOK, followStatusResponse{Following: false, Followers: followers})
	}
}

func IsFollowing(repo repository.FollowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		followerID := uidAny.(string)
		targetID := c.Param("id")
		following, err := repo.IsFollowing(c.Request.Context(), followerID, targetID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, followStatusResponse{Following: following})
	}
}

func GetFollowers(repo repository.FollowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetID := c.Param("id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		users, err := repo.GetFollowers(c.Request.Context(), targetID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	}
}

func GetFollowing(repo repository.FollowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.Param("id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		users, err := repo.GetFollowing(c.Request.Context(), uid, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	}
}

func GetFollowCounts(repo repository.FollowRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.Param("id")
		followers, err1 := repo.CountFollowers(c.Request.Context(), uid)
		following, err2 := repo.CountFollowing(c.Request.Context(), uid)
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get counts"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"followers": followers, "following": following})
	}
}
