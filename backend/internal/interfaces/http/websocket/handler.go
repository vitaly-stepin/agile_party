package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/vitaly-stepin/agile_party/internal/application"
	"github.com/vitaly-stepin/agile_party/internal/application/dto"
)

// Handler handles WebSocket connections
type Handler struct {
	hub           *Hub
	roomService   *application.RoomService
	userService   *application.UserService
	votingService *application.VotingService
}

// NewHandler creates a new WebSocket handler
func NewHandler(
	hub *Hub,
	roomService *application.RoomService,
	userService *application.UserService,
	votingService *application.VotingService,
) *Handler {
	return &Handler{
		hub:           hub,
		roomService:   roomService,
		userService:   userService,
		votingService: votingService,
	}
}

// HandleConnection handles WebSocket connection for a room
// Route: WS /ws/rooms/:id?userId=xxx&userName=xxx
func (h *Handler) HandleConnection(c *websocket.Conn) {
	roomID := c.Params("id")
	userID := c.Query("userId")
	userName := c.Query("userName")

	// Validate parameters
	if roomID == "" || userID == "" || userName == "" {
		sendErrorAndClose(c, "Room ID, User ID, and User Name are required", "MISSING_PARAMS")
		return
	}

	// Create context for operations
	ctx := context.Background()

	// Verify room exists
	_, err := h.roomService.GetRoom(ctx, roomID)
	if err != nil {
		sendErrorAndClose(c, "Room not found", "ROOM_NOT_FOUND")
		return
	}

	// Add user to room
	user, err := h.userService.JoinRoom(ctx, roomID, userID, userName)
	if err != nil {
		sendErrorAndClose(c, err.Error(), "JOIN_ERROR")
		return
	}

	// Create client
	client := NewClient(h.hub, c, roomID, userID, user.Name, h.userService, h.votingService)

	// Register client with hub
	h.hub.register <- client

	// Send initial room state to the new client
	if err := h.sendInitialState(client); err != nil {
		log.Printf("Error sending initial state: %v", err)
		h.hub.unregister <- client
		return
	}

	// Broadcast user joined event to other clients
	h.broadcastUserJoined(roomID, user)

	// Start client read/write pumps
	go client.writePump()
	client.readPump() // Blocking call
}

// sendInitialState sends the current room state to a newly connected client
func (h *Handler) sendInitialState(client *Client) error {
	ctx := context.Background()
	state, err := h.roomService.GetRoomState(ctx, client.roomID)
	if err != nil {
		return err
	}

	// Convert users to payload format
	users := make([]UserPayload, 0, len(state.Users))
	for _, u := range state.Users {
		users = append(users, UserPayload{
			UserID:   u.UserID,
			Name:     u.Name,
			IsVoted:  u.IsVoted,
			IsOnline: u.IsOnline,
		})
	}

	// Convert votes to payload format (only if revealed)
	var votes []VotePayload2
	if state.IsRevealed && len(state.Votes) > 0 {
		votes = make([]VotePayload2, 0, len(state.Votes))
		for _, v := range state.Votes {
			votes = append(votes, VotePayload2{
				UserID: v.UserID,
				Value:  v.Value,
			})
		}
	}

	event := Event{
		Type: EventTypeRoomState,
		Payload: RoomStatePayload{
			RoomID:     state.RoomID,
			RoomName:   state.RoomName,
			Users:      users,
			Votes:      votes,
			IsRevealed: state.IsRevealed,
			Average:    state.Average,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	select {
	case client.send <- data:
		return nil
	default:
		return fiber.ErrRequestTimeout
	}
}

// broadcastUserJoined sends user joined event to all clients in the room
func (h *Handler) broadcastUserJoined(roomID string, user *dto.UserResponse) {
	event := Event{
		Type: EventTypeUserJoined,
		Payload: UserJoinedPayload{
			UserID: user.UserID,
			Name:   user.Name,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling user joined event: %v", err)
		return
	}

	h.hub.BroadcastToRoom(roomID, data)
}

// sendErrorAndClose sends an error message and closes the connection
func sendErrorAndClose(c *websocket.Conn, message, code string) {
	event := Event{
		Type: EventTypeError,
		Payload: ErrorPayload{
			Message: message,
			Code:    code,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling error event: %v", err)
		c.Close()
		return
	}

	if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error writing error message: %v", err)
	}

	c.Close()
}
