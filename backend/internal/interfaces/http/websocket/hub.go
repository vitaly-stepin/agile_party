package websocket

import (
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by room ID
	rooms map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients in a room
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// BroadcastMessage contains the message and target room
type BroadcastMessage struct {
	RoomID  string
	Message []byte
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
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
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToRoom(message)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.roomID] == nil {
		h.rooms[client.roomID] = make(map[*Client]bool)
	}
	h.rooms[client.roomID][client] = true
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.roomID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)

			// Remove room if empty
			if len(clients) == 0 {
				delete(h.rooms, client.roomID)
			}
		}
	}
}

// broadcastToRoom sends a message to all clients in a room
func (h *Hub) broadcastToRoom(message *BroadcastMessage) {
	h.mu.RLock()
	clients := h.rooms[message.RoomID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.send <- message.Message:
		default:
			// Client's send channel is full, close it
			h.mu.Lock()
			close(client.send)
			delete(h.rooms[message.RoomID], client)
			h.mu.Unlock()
		}
	}
}

// BroadcastToRoom sends a message to all clients in a specific room
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: message,
	}
}

// GetRoomClientCount returns the number of connected clients in a room
func (h *Hub) GetRoomClientCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		return len(clients)
	}
	return 0
}
