package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type WsHub struct {
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	RoomID  string
	Message WsMessage
	Exclude *Client // Optional: exclude this client from broadcast
}

func NewHub() *WsHub {
	return &WsHub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage, 256),
	}
}

func (h *WsHub) Run() {
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

func (h *WsHub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.RoomID] == nil {
		h.rooms[client.RoomID] = make(map[*Client]bool)
	}
	h.rooms[client.RoomID][client] = true

	log.Printf("Client %s registered to room %s. Total clients in room: %d",
		client.UserID, client.RoomID, len(h.rooms[client.RoomID]))
}

func (h *WsHub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.RoomID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)

			// Clean up empty rooms, consider refactoring because might it may cause unexpected behavior
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

func (h *WsHub) broadcastToRoom(msg BroadcastMessage) {
	h.mu.RLock()
	clients := h.rooms[msg.RoomID]
	h.mu.RUnlock()

	if clients == nil {
		return
	}

	data, err := json.Marshal(msg.Message)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	for client := range clients {
		if msg.Exclude != nil && client == msg.Exclude {
			continue
		}

		select {
		case client.send <- data:
		default:
			// Client's send channel is full, close connection
			log.Printf("Client %s send channel full, closing connection", client.UserID)
			h.unregisterClient(client)
		}
	}
}

func (h *WsHub) BroadcastToRoom(roomID string, message WsMessage, exclude *Client) {
	h.broadcast <- BroadcastMessage{
		RoomID:  roomID,
		Message: message,
		Exclude: exclude,
	}
}

func (h *WsHub) GetRoomClientCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		return len(clients)
	}
	return 0
}
