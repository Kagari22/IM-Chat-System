package chat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"IM_Chat_System/internal/auth"
	"IM_Chat_System/internal/model"
	"IM_Chat_System/internal/mq"
	"IM_Chat_System/internal/presence"
	"IM_Chat_System/internal/ratelimit"
	"IM_Chat_System/internal/service"
	"IM_Chat_System/internal/tokenblacklist"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 70 * time.Second
	pingPeriod = 30 * time.Second
)

type Hub struct {
	mu        sync.RWMutex
	clients   map[int64]*Client
	secret    string
	nodeID    string
	messages  *service.MessageService
	presence  presence.Store
	blacklist tokenblacklist.Store
	limiter   ratelimit.Store
}

type Client struct {
	userID int64
	conn   *websocket.Conn
	send   chan any
	hub    *Hub
}

type IncomingMessage struct {
	Type    string `json:"type"`
	To      int64  `json:"to"`
	Content string `json:"content"`
}

type OutgoingMessage struct {
	Type    string        `json:"type"`
	Message model.Message `json:"message,omitempty"`
	Error   string        `json:"error,omitempty"`
}

func NewHub(messages *service.MessageService, secret, nodeID string, presenceStore presence.Store, blacklist tokenblacklist.Store, limiter ratelimit.Store) *Hub {
	return &Hub{
		clients:   make(map[int64]*Client),
		secret:    secret,
		nodeID:    nodeID,
		messages:  messages,
		presence:  presenceStore,
		blacklist: blacklist,
		limiter:   limiter,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if h.blacklist != nil {
		blocked, err := h.blacklist.Contains(r.Context(), token)
		if err != nil {
			http.Error(w, "check token blacklist failed", http.StatusInternalServerError)
			return
		}
		if blocked {
			http.Error(w, "token has been revoked", http.StatusUnauthorized)
			return
		}
	}
	claims, err := auth.ParseToken(h.secret, token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket upgrade:", err)
		return
	}

	client := &Client{
		userID: claims.UserID,
		conn:   conn,
		send:   make(chan any, 16),
		hub:    h,
	}

	h.register(client)
	go client.writePump()
	go client.readPump()
}

func (h *Hub) register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if old := h.clients[client.userID]; old != nil {
		old.conn.Close()
	}
	h.clients[client.userID] = client
	if h.presence != nil {
		if err := h.presence.SetOnline(context.Background(), client.userID, h.nodeID); err != nil {
			log.Println("presence set online:", err)
		}
	}
	log.Printf("user %d online\n", client.userID)
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.userID] == client {
		delete(h.clients, client.userID)
		close(client.send)
		if h.presence != nil {
			if err := h.presence.SetOffline(context.Background(), client.userID); err != nil {
				log.Println("presence set offline:", err)
			}
		}
		log.Printf("user %d offline\n", client.userID)
	}
}

func (h *Hub) deliver(userID int64, payload any) bool {
	h.mu.RLock()
	client := h.clients[userID]
	h.mu.RUnlock()

	if client == nil {
		return false
	}

	select {
	case client.send <- payload:
		return true
	default:
		client.conn.Close()
		return false
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var input IncomingMessage
		if err := c.conn.ReadJSON(&input); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("websocket read:", err)
			}
			return
		}

		if input.Type != "chat" {
			c.send <- OutgoingMessage{Type: "error", Error: "unsupported message type"}
			continue
		}
		if c.hub.limiter != nil {
			allowed, err := c.hub.limiter.Allow(context.Background(), messageRateKey(c.userID), 20, time.Minute)
			if err != nil {
				c.send <- OutgoingMessage{Type: "error", Error: "rate limit check failed"}
				continue
			}
			if !allowed {
				c.send <- OutgoingMessage{Type: "error", Error: "too many messages, slow down"}
				continue
			}
		}

		message, err := c.hub.messages.SaveText(context.Background(), c.userID, input.To, input.Content)
		if err != nil {
			c.send <- OutgoingMessage{Type: "error", Error: err.Error()}
			continue
		}

		c.send <- OutgoingMessage{Type: "ack", Message: message}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case payload, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(payload)
			if err != nil {
				log.Println("json marshal:", err)
				continue
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			if c.hub.presence != nil {
				if err := c.hub.presence.SetOnline(context.Background(), c.userID, c.hub.nodeID); err != nil {
					log.Println("presence heartbeat:", err)
				}
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func messageRateKey(userID int64) string {
	return "ratelimit:ws:message:user:" + strconv.FormatInt(userID, 10)
}

func (h *Hub) DispatchMessageCreated(ctx context.Context, event mq.MessageCreatedEvent) error {
	if h.presence != nil {
		nodeID, online, err := h.presence.GetOnlineNode(ctx, event.ToUserID)
		if err != nil {
			return err
		}
		if !online || nodeID != h.nodeID {
			return nil
		}
	}

	payload := OutgoingMessage{
		Type: "chat",
		Message: model.Message{
			ID:          event.MessageID,
			FromUserID:  event.FromUserID,
			ToUserID:    event.ToUserID,
			ContentType: event.ContentType,
			Content:     event.Content,
			ObjectKey:   event.ObjectKey,
			ObjectURL:   event.ObjectURL,
			FileName:    event.FileName,
			FileSize:    event.FileSize,
			CreatedAt:   event.CreatedAt,
		},
	}
	h.deliver(event.ToUserID, payload)
	return nil
}
