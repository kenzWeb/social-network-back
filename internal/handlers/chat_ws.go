package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	imodels "modern-social-media/internal/models"
	"modern-social-media/internal/repository"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type WSMessagePayload struct {
	ConversationID string `json:"conversation_id"`
	Body           string `json:"body"`
}

type WSTypingPayload struct {
	ConversationID string `json:"conversation_id"`
	IsTyping       bool   `json:"is_typing"`
}

type WSReadPayload struct {
	ConversationID string `json:"conversation_id"`
}

type presenceUpdate struct {
	UserID   string `json:"user_id"`
	Online   bool   `json:"online"`
	LastSeen int64  `json:"last_seen"`
}

type connection struct {
	userID string
	conn   *websocket.Conn
	send   chan []byte
}

type Hub struct {
	mu           sync.RWMutex
	connections  map[string]map[*connection]struct{}
	onlineStatus map[string]bool
	lastSeen     map[string]time.Time
}

func NewHub() *Hub {
	return &Hub{
		connections:  make(map[string]map[*connection]struct{}),
		onlineStatus: make(map[string]bool),
		lastSeen:     make(map[string]time.Time),
	}
}

func (h *Hub) addConn(userID string, c *connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.connections[userID] == nil {
		h.connections[userID] = make(map[*connection]struct{})
	}
	h.connections[userID][c] = struct{}{}
	h.onlineStatus[userID] = true
}

func (h *Hub) removeConn(userID string, c *connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.connections[userID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.connections, userID)
			h.onlineStatus[userID] = false
			h.lastSeen[userID] = time.Now()
		}
	}
}

func (h *Hub) broadcastToUsers(userIDs []string, payload any) {
	b, _ := json.Marshal(payload)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, uid := range userIDs {
		if conns, ok := h.connections[uid]; ok {
			for c := range conns {
				select {
				case c.send <- b:
				default:
				}
			}
		}
	}
}

func (h *Hub) sendToUser(userID string, payload any) {
	h.broadcastToUsers([]string{userID}, payload)
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type ChatWSDeps struct {
	Models    repository.Models
	JWTSecret string
	Hub       *Hub
}

func ChatWSHandler(deps ChatWSDeps) gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenStr := c.Query("token")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			return
		}

		parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(deps.JWTSecret), nil
		})
		if err != nil || !parsed.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok || claims["type"] != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
			return
		}

		userID, _ := claims["sub"].(string)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid subject"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		client := &connection{userID: userID, conn: conn, send: make(chan []byte, 256)}
		deps.Hub.addConn(userID, client)

		deps.Hub.sendToUser(userID, WSEvent{Type: "presence", Data: mustJSON(presenceUpdate{UserID: userID, Online: true, LastSeen: time.Now().Unix()})})

		go writer(client)
		reader(c, deps, client)
	}
}

func writer(cn *connection) {
	defer func() { _ = cn.conn.Close() }()
	for msg := range cn.send {
		cn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := cn.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func reader(ctx *gin.Context, deps ChatWSDeps, cn *connection) {
	defer func() {
		deps.Hub.removeConn(cn.userID, cn)
		close(cn.send)
		_ = cn.conn.Close()
		deps.Hub.sendToUser(cn.userID, WSEvent{Type: "presence", Data: mustJSON(presenceUpdate{UserID: cn.userID, Online: false, LastSeen: time.Now().Unix()})})
	}()

	cn.conn.SetReadLimit(1 << 20)
	cn.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	cn.conn.SetPongHandler(func(string) error {
		cn.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := cn.conn.ReadMessage()
		if err != nil {
			log.Println("ws read:", err)
			return
		}
		var evt WSEvent
		if err := json.Unmarshal(data, &evt); err != nil {
			deps.Hub.sendToUser(cn.userID, WSEvent{Type: "error", Data: mustJSON(gin.H{"error": "bad_event"})})
			continue
		}
		switch evt.Type {
		case "typing":
			var p WSTypingPayload
			if json.Unmarshal(evt.Data, &p) == nil {
				others := getConversationPeers(ctx, deps.Models, p.ConversationID, cn.userID)
				deps.Hub.broadcastToUsers(others, WSEvent{Type: "typing", Data: mustJSON(gin.H{"conversation_id": p.ConversationID, "user_id": cn.userID, "is_typing": p.IsTyping})})
			}
		case "message":
			var p WSMessagePayload
			if json.Unmarshal(evt.Data, &p) == nil && p.Body != "" {
				msg := &imodels.Message{ConversationID: p.ConversationID, SenderID: cn.userID, Body: p.Body}
				if err := deps.Models.Chat.CreateMessage(ctx.Request.Context(), msg); err != nil {
					deps.Hub.sendToUser(cn.userID, WSEvent{Type: "error", Data: mustJSON(gin.H{"error": "save_failed"})})
					continue
				}
				peers := getConversationPeers(ctx, deps.Models, p.ConversationID, "")
				deps.Hub.broadcastToUsers(peers, WSEvent{Type: "message", Data: mustJSON(gin.H{"conversation_id": p.ConversationID, "sender_id": cn.userID, "body": p.Body, "created_at": time.Now().Unix()})})
			}
		case "read":
			var p WSReadPayload
			if json.Unmarshal(evt.Data, &p) == nil {
				_ = deps.Models.Chat.UpdateLastRead(ctx.Request.Context(), p.ConversationID, cn.userID, time.Now())
			}
		default:
			deps.Hub.sendToUser(cn.userID, WSEvent{Type: "error", Data: mustJSON(gin.H{"error": "unknown_type"})})
		}
	}
}

func getConversationPeers(ctx *gin.Context, repos repository.Models, conversationID, excludeUser string) []string {
	type row struct{ UserID string }
	var rows []row
	_ = repos.DB.WithContext(ctx.Request.Context()).
		Raw("SELECT user_id FROM conversation_participants WHERE conversation_id = ?", conversationID).
		Scan(&rows).Error
	res := make([]string, 0, len(rows))
	for _, r := range rows {
		if excludeUser != "" && r.UserID == excludeUser {
			continue
		}
		res = append(res, r.UserID)
	}
	return res
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func ListConversations(repos repository.Models) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		items, err := repos.Chat.ListUserConversations(c.Request.Context(), userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"conversations": items})
	}
}

func ListMessages(repos repository.Models) gin.HandlerFunc {
	return func(c *gin.Context) {
		convID := c.Param("id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		items, err := repos.Chat.ListMessages(c.Request.Context(), convID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
			return
		}

		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
		c.JSON(http.StatusOK, gin.H{"messages": items})
	}
}

type sendDirectBody struct {
	Body string `json:"body"`
}

func SendDirectMessage(repos repository.Models, hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		from := c.GetString("userID")
		to := c.Param("user_id")
		var req sendDirectBody
		if err := c.BindJSON(&req); err != nil || req.Body == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_body"})
			return
		}
		conv, err := repos.Chat.GetOrCreateDirectConversation(c.Request.Context(), from, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "conv_failed"})
			return
		}
		msg := &imodels.Message{ConversationID: conv.ID, SenderID: from, Body: req.Body}
		if err := repos.Chat.CreateMessage(c.Request.Context(), msg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "save_failed"})
			return
		}
		hub.broadcastToUsers([]string{from, to}, WSEvent{Type: "message", Data: mustJSON(gin.H{"conversation_id": conv.ID, "sender_id": from, "body": req.Body, "created_at": time.Now().Unix()})})
		c.JSON(http.StatusOK, gin.H{"message": msg})
	}
}

func MarkRead(repos repository.Models) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		convID := c.Param("id")
		_ = repos.Chat.UpdateLastRead(c.Request.Context(), convID, userID, time.Now())
		c.Status(http.StatusNoContent)
	}
}

func GetPresence(h *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.Param("user_id")
		h.mu.RLock()
		online := h.onlineStatus[uid]
		last := h.lastSeen[uid]
		h.mu.RUnlock()
		var lastUnix int64
		if !last.IsZero() {
			lastUnix = last.Unix()
		}
		c.JSON(http.StatusOK, gin.H{"user_id": uid, "online": online, "last_seen": lastUnix})
	}
}
