package handlers

import (
	"net/http"
	"strconv"

	"modern-social-media/internal/repository"

	"github.com/gin-gonic/gin"
)

// @name NotificationResponse
type notificationResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Read      bool      `json:"read"`
	TargetID  *string   `json:"target_id,omitempty"`
	CreatedAt string    `json:"created_at"`
	Actor     actorInfo `json:"actor"`
}

// @name ActorInfo
type actorInfo struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	AvatarURL  string `json:"avatar_url"`
	IsVerified bool   `json:"is_verified"`
}

// @name NotificationListResponse
type notificationListResponse struct {
	Notifications []notificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
}

// @Summary Get notifications
// @Description Get all notifications for the authenticated user
// @Tags notifications
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} notificationListResponse
// @Security BearerAuth
// @Router /notifications [get]
func GetNotifications(repo repository.NotificationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		notifications, err := repo.GetByUserID(c.Request.Context(), userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notifications"})
			return
		}

		unreadCount, _ := repo.CountUnread(c.Request.Context(), userID)

		response := make([]notificationResponse, len(notifications))
		for i, n := range notifications {
			response[i] = notificationResponse{
				ID:        n.ID,
				Type:      string(n.Type),
				Read:      n.Read,
				TargetID:  n.TargetID,
				CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z"),
				Actor: actorInfo{
					ID:         n.Actor.ID,
					Username:   n.Actor.Username,
					FirstName:  n.Actor.FirstName,
					LastName:   n.Actor.LastName,
					AvatarURL:  n.Actor.AvatarURL,
					IsVerified: n.Actor.IsVerified,
				},
			}
		}

		c.JSON(http.StatusOK, notificationListResponse{
			Notifications: response,
			UnreadCount:   unreadCount,
		})
	}
}

// @Summary Get unread notifications
// @Description Get unread notifications for the authenticated user
// @Tags notifications
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} notificationListResponse
// @Security BearerAuth
// @Router /notifications/unread [get]
func GetUnreadNotifications(repo repository.NotificationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		notifications, err := repo.GetUnreadByUserID(c.Request.Context(), userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notifications"})
			return
		}

		response := make([]notificationResponse, len(notifications))
		for i, n := range notifications {
			response[i] = notificationResponse{
				ID:        n.ID,
				Type:      string(n.Type),
				Read:      n.Read,
				TargetID:  n.TargetID,
				CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z"),
				Actor: actorInfo{
					ID:         n.Actor.ID,
					Username:   n.Actor.Username,
					FirstName:  n.Actor.FirstName,
					LastName:   n.Actor.LastName,
					AvatarURL:  n.Actor.AvatarURL,
					IsVerified: n.Actor.IsVerified,
				},
			}
		}

		c.JSON(http.StatusOK, gin.H{"notifications": response})
	}
}

// @Summary Get unread count
// @Description Get the count of unread notifications
// @Tags notifications
// @Produce json
// @Success 200 {object} map[string]int64
// @Security BearerAuth
// @Router /notifications/unread/count [get]
func GetUnreadCount(repo repository.NotificationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		count, err := repo.CountUnread(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread count"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"unread_count": count})
	}
}

// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /notifications/{id}/read [patch]
func MarkNotificationAsRead(repo repository.NotificationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)
		notificationID := c.Param("id")

		if err := repo.MarkAsRead(c.Request.Context(), notificationID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark as read"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
	}
}

// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the authenticated user
// @Tags notifications
// @Produce json
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /notifications/read-all [patch]
func MarkAllNotificationsAsRead(repo repository.NotificationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)

		if err := repo.MarkAllAsRead(c.Request.Context(), userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all as read"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "all marked as read"})
	}
}

// @Summary Delete notification
// @Description Delete a specific notification
// @Tags notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /notifications/{id} [delete]
func DeleteNotification(repo repository.NotificationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidAny, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID := uidAny.(string)
		notificationID := c.Param("id")

		if err := repo.Delete(c.Request.Context(), notificationID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notification"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "notification deleted"})
	}
}
