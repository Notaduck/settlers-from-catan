package hub

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

// Client represents a WebSocket client
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	PlayerID  string
	GameID    string
	OnMessage func([]byte)
	closed    bool
	mu        sync.RWMutex
}

// NewClient creates a new client
func NewClient(hub *Hub, conn *websocket.Conn, playerID, gameID string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		PlayerID: playerID,
		GameID:   gameID,
	}
}

// ReadPump pumps messages from the WebSocket to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		if c.OnMessage != nil {
			c.OnMessage(message)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send sends a message to the client
func (c *Client) Send(message []byte) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		log.Printf("Attempted to send to closed client %s", c.PlayerID)
		return
	}

	select {
	case c.send <- message:
	default:
		log.Printf("Client send buffer full for player %s", c.PlayerID)
	}
}

// Close marks the client as closed
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}

// IsClosed returns whether the client is closed
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}
