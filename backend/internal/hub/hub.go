package hub

import (
	"encoding/json"
	"log"
	"sync"
)

// Message represents a WebSocket message
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	games      map[string]map[*Client]bool // gameID -> clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

// BroadcastMessage is a message to broadcast to a game
type BroadcastMessage struct {
	GameID  string
	Message []byte
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		games:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if client.GameID != "" {
				if h.games[client.GameID] == nil {
					h.games[client.GameID] = make(map[*Client]bool)
				}
				h.games[client.GameID][client] = true
			}
			h.mu.Unlock()
			log.Printf("Client registered: %s for game %s", client.PlayerID, client.GameID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.GameID != "" {
					delete(h.games[client.GameID], client)
				}
				client.Close()
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s", client.PlayerID)

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.games[message.GameID]; ok {
				for client := range clients {
					select {
					case client.send <- message.Message:
					default:
						client.Close()
						close(client.send)
						delete(h.clients, client)
						delete(h.games[message.GameID], client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToGame sends a message to all clients in a game
func (h *Hub) BroadcastToGame(gameID string, message []byte) {
	h.broadcast <- &BroadcastMessage{
		GameID:  gameID,
		Message: message,
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
