package chat

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/muah1987/Aihub/internal/models"
)

type Client struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
	Conn      *websocket.Conn
	Send      chan []byte
}

type Hub struct {
	// projectID -> set of clients
	rooms      map[uuid.UUID]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	ProjectID uuid.UUID
	Message   *models.Message
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.ProjectID]; !ok {
				h.rooms[client.ProjectID] = make(map[*Client]bool)
			}
			h.rooms[client.ProjectID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.ProjectID]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.Send)
					if len(room) == 0 {
						delete(h.rooms, client.ProjectID)
					}
				}
			}
			h.mu.Unlock()

		case bm := <-h.broadcast:
			h.mu.RLock()
			if room, ok := h.rooms[bm.ProjectID]; ok {
				data, err := json.Marshal(bm.Message)
				if err != nil {
					log.Printf("Failed to marshal message: %v", err)
					h.mu.RUnlock()
					continue
				}
				for client := range room {
					select {
					case client.Send <- data:
					default:
						close(client.Send)
						delete(room, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(projectID uuid.UUID, msg *models.Message) {
	h.broadcast <- &BroadcastMessage{
		ProjectID: projectID,
		Message:   msg,
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
