package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub manages WebSocket connections and broadcasts messages
type Hub struct {
	// Registered clients by room
	rooms map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to specific room
	broadcast chan BroadcastMessage

	// Mutex to protect rooms map
	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast to a room
type BroadcastMessage struct {
	RoomID  string
	Message Message
	Exclude *Client // Optional: exclude this client from broadcast
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToRoom(message)
		}
	}
}

// registerClient adds a client to a room
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.RoomID] == nil {
		h.rooms[client.RoomID] = make(map[*Client]bool)
	}
	h.rooms[client.RoomID][client] = true

	log.Printf("Client %s registered to room %s. Total clients in room: %d",
		client.UserID, client.RoomID, len(h.rooms[client.RoomID]))
}

// unregisterClient removes a client from a room
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.RoomID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)

			// Clean up empty rooms
			if len(clients) == 0 {
				delete(h.rooms, client.RoomID)
				log.Printf("Room %s is now empty and removed", client.RoomID)
			} else {
				log.Printf("Client %s unregistered from room %s. Remaining clients: %d",
					client.UserID, client.RoomID, len(clients))
			}
		}
	}
}

// broadcastToRoom sends a message to all clients in a room
func (h *Hub) broadcastToRoom(msg BroadcastMessage) {
	h.mu.RLock()
	clients := h.rooms[msg.RoomID]
	h.mu.RUnlock()

	if clients == nil {
		return
	}

	// Marshal message once
	data, err := json.Marshal(msg.Message)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	// Send to all clients in room
	for client := range clients {
		// Skip excluded client (e.g., the sender)
		if msg.Exclude != nil && client == msg.Exclude {
			continue
		}

		select {
		case client.send <- data:
			// Message sent successfully
		default:
			// Client's send channel is full, close connection
			log.Printf("Client %s send channel full, closing connection", client.UserID)
			h.unregisterClient(client)
		}
	}
}

// BroadcastToRoom queues a message for broadcast to a room
func (h *Hub) BroadcastToRoom(roomID string, message Message, exclude *Client) {
	h.broadcast <- BroadcastMessage{
		RoomID:  roomID,
		Message: message,
		Exclude: exclude,
	}
}

// GetRoomClientCount returns the number of clients in a room
func (h *Hub) GetRoomClientCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		return len(clients)
	}
	return 0
}
