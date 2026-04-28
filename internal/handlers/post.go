package handlers

import (
	"fmt"
	"modern-social-media/internal/models"
	"modern-social-media/internal/repository"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// @name PostResponse
type PostResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url"`
	Likes     int       `json:"likes_count"`
	Comments  int       `json:"comments_count"`
	CreatedAt time.Time `json:"created_at"`
}

// @Summary Get user posts
// @Description Get posts by user
// @Tags posts
// @Produce json
// @Success 200 {array} PostResponse
// @Router /posts [get]
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

		var response []PostResponse
		for _, p := range posts {
			response = append(response, PostResponse{
				ID:        p.ID,
				UserID:    p.UserID,
				Content:   p.Content,
				ImageURL:  p.ImageURL,
				Likes:     p.LikesCount,
				Comments:  p.CommentsCount,
				CreatedAt: p.CreatedAt,
			})
		}
		c.JSON(http.StatusOK, response)
	}
}


// @Summary Get all posts
// @Description Get posts post all users
// @Tags posts
// @Produce json
// @Success 200 {array} PostResponse
// @Router /posts/all [get]
func GetAllPosts(postRepo repository.PostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, err := postRepo.GetAllPosts(c.Request.Context())

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
			return
		}

		c.JSON(http.StatusOK, posts)
	}
}

// @name CreatePostRequest
type CreatePostRequest struct {
	Content  string `json:"content"`
	ImageURL string `json:"imageUrl"`
}

// @name UpdatePostRequest
type UpdatePostRequest struct {
	Content  string `json:"content"`
	ImageURL string `json:"imageUrl"`
}

// @Summary Create post
// @Description Create a new post with optional image
// @Tags posts
// @Accept multipart/form-data
// @Produce json
// @Param content formData string true "Post content"
// @Param image formData file false "Post image"
// @Success 201 {object} PostResponse
// @Router /posts [post]
func CreatePost(postRepo repository.PostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		const maxUploadSize = 5 << 20
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

		if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request too large (max 5MB)"})
			return
		}

		content := c.PostForm("content")
		if strings.TrimSpace(content) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
			return
		}

		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidAny.(string)

		var imageURL string
		file, err := c.FormFile("image")
		if err == nil {
			ext := strings.ToLower(filepath.Ext(filepath.Base(file.Filename)))
			
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Only image files are allowed (.jpg, .png, .gif, .webp)"})
				return
			}

			os.MkdirAll("uploads/posts", os.ModePerm)

			filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			path := filepath.Join("uploads", "posts", filename)

			if err := c.SaveUploadedFile(file, path); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
				return
			}
			imageURL = "/uploads/posts/" + filename
		}

		post := &models.Post{
			UserID:   userID,
			Content:  content,
			ImageURL: imageURL,
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
			c.JSON(http.StatusCreated, PostResponse{
				ID:        post.ID,
				UserID:    post.UserID,
				Content:   post.Content,
				ImageURL:  post.ImageURL,
				CreatedAt: post.CreatedAt,
			})
			return
		}
		
		c.JSON(http.StatusCreated, PostResponse{
			ID:        full.ID,
			UserID:    full.UserID,
			Content:   full.Content,
			ImageURL:  full.ImageURL,
			Likes:     full.LikesCount,
			Comments:  full.CommentsCount,
			CreatedAt: full.CreatedAt,
		})
	}
}

// @Summary Update post
// @Description Update a post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body UpdatePostRequest true "Post content"
// @Success 200 {object} PostResponse
// @Router /posts/{id} [put]
func UpdatePost(postRepo repository.PostRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req UpdatePostRequest
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
			c.JSON(http.StatusOK, PostResponse{
				ID:        post.ID,
				UserID:    post.UserID,
				Content:   post.Content,
				ImageURL:  post.ImageURL,
				CreatedAt: post.CreatedAt,
			})
			return
		}
		
		c.JSON(http.StatusOK, PostResponse{
			ID:        full.ID,
			UserID:    full.UserID,
			Content:   full.Content,
			ImageURL:  full.ImageURL,
			Likes:     full.LikesCount,
			Comments:  full.CommentsCount,
			CreatedAt: full.CreatedAt,
		})
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
