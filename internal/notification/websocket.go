package notification

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应限制 origin
	},
}

type Client struct {
	userID string
	conn   *websocket.Conn
	send   chan []byte
}

type Hub struct {
	clients    map[string][]*Client
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

var hub = &Hub{
	clients:    make(map[string][]*Client),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func init() {
	go hub.run()
	BroadcastNotification = broadcastToClients
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.userID] = append(h.clients[client.userID], client)
			h.mutex.Unlock()
			log.Printf("WebSocket client registered for user: %s", client.userID)

		case client := <-h.unregister:
			h.mutex.Lock()
			for i, c := range h.clients[client.userID] {
				if c == client {
					h.clients[client.userID] = append(h.clients[client.userID][:i], h.clients[client.userID][i+1:]...)
					close(c.send)
					break
				}
			}
			h.mutex.Unlock()
			log.Printf("WebSocket client unregistered for user: %s", client.userID)
		}
	}
}

// broadcastToClients 广播通知给指定用户
func broadcastToClients(userID string, n *Notification) {
	hub.mutex.RLock()
	clients := hub.clients[userID]
	hub.mutex.RUnlock()

	msg, _ := json.Marshal(map[string]interface{}{
		"type": "notification",
		"data": n,
	})

	for _, client := range clients {
		select {
		case client.send <- msg:
		default:
			close(client.send)
		}
	}
}

func HandleWebSocket(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		userID: userID.(string),
		conn:   conn,
		send:   make(chan []byte, 256),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}