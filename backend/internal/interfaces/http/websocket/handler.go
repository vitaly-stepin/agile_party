package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"github.com/vitaly-stepin/agile_party/internal/application"
	"github.com/vitaly-stepin/agile_party/internal/application/dto"
)

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		// Allow all origins for MVP (configure properly in production)
		return true
	},
}

// Handler manages WebSocket connections and message routing
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

// HandleConnection handles WebSocket connection upgrades
func (h *Handler) HandleConnection(c *fiber.Ctx) error {
	roomID := c.Params("id")
	userID := c.Query("userId")
	nickname := c.Query("nickname")

	log.Printf("[DEBUG] WebSocket connection - roomID: '%s', userID: '%s', path: '%s'", roomID, userID, c.Path())

	if roomID == "" || userID == "" || nickname == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "roomId, userId, and nickname are required",
		})
	}

	// Upgrade the connection
	err := upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
		h.handleWebSocket(conn, roomID, userID, nickname)
	})

	return err
}

// handleWebSocket manages the WebSocket lifecycle for a client
func (h *Handler) handleWebSocket(conn *websocket.Conn, roomID, userID, nickname string) {
	ctx := context.Background()

	// Join room (adds user to in-memory state)
	if err := h.userService.JoinRoom(ctx, roomID, userID, nickname); err != nil {
		log.Printf("Failed to join room %s for user %s: %v", roomID, userID, err)
		conn.WriteJSON(Message{
			Type: EventTypeError,
			Payload: ErrorPayload{
				Message: "Failed to join room: " + err.Error(),
				Code:    "JOIN_FAILED",
			},
		})
		conn.Close()
		return
	}

	// Create client and register with hub
	log.Printf("[DEBUG] Creating client - roomID: '%s', userID: '%s'", roomID, userID)
	client := NewClient(conn, h.hub, roomID, userID, h)
	log.Printf("[DEBUG] Client created - client.RoomID: '%s', client.UserID: '%s'", client.RoomID, client.UserID)
	h.hub.register <- client

	// Send initial room state to the new client
	if err := h.sendRoomState(client); err != nil {
		log.Printf("Failed to send initial state to user %s in room %s: %v", userID, roomID, err)
	}

	// Broadcast user joined to other clients
	h.hub.BroadcastToRoom(roomID, Message{
		Type: EventTypeUserJoined,
		Payload: UserJoinedPayload{
			UserID:   userID,
			Nickname: nickname,
		},
	}, client)

	// Handle disconnection cleanup
	defer func() {
		// Remove user from room
		if err := h.userService.LeaveRoom(ctx, roomID, userID); err != nil {
			log.Printf("Failed to remove user %s from room %s: %v", userID, roomID, err)
		}

		// Broadcast user left
		h.hub.BroadcastToRoom(roomID, Message{
			Type: EventTypeUserLeft,
			Payload: UserLeftPayload{
				UserID: userID,
			},
		}, nil)

		log.Printf("User %s disconnected from room %s", userID, roomID)
	}()

	// Start client pumps
	client.Start()
}

// HandleMessage processes incoming WebSocket messages (implements MessageHandler)
func (h *Handler) HandleMessage(client *Client, msg Message) error {
	ctx := context.Background()

	switch msg.Type {
	case EventTypeVote:
		return h.handleVote(ctx, client, msg)

	case EventTypeReveal:
		return h.handleReveal(ctx, client)

	case EventTypeClear:
		return h.handleClear(ctx, client)

	case EventTypeUpdateNickname:
		return h.handleUpdateNickname(ctx, client, msg)

	case EventTypeSetTask:
		return h.handleSetTask(ctx, client, msg)

	default:
		return fmt.Errorf("unknown event type: %s", msg.Type)
	}
}

// handleVote processes a vote submission
func (h *Handler) handleVote(ctx context.Context, client *Client, msg Message) error {
	var payload VotePayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid vote payload: %w", err)
	}

	// Submit vote
	if err := h.votingService.SubmitVote(ctx, client.RoomID, client.UserID, payload.Value); err != nil {
		return fmt.Errorf("failed to submit vote: %w", err)
	}

	// Get updated room state
	roomState, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	// Broadcast updated room state to all clients
	h.hub.BroadcastToRoom(client.RoomID, Message{
		Type:    EventTypeRoomState,
		Payload: h.convertRoomStateToPayload(roomState),
	}, nil)

	return nil
}

// handleReveal processes a reveal request
func (h *Handler) handleReveal(ctx context.Context, client *Client) error {
	// Reveal votes
	result, err := h.votingService.RevealVotes(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to reveal votes: %w", err)
	}

	// Get room state to access user names
	state, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	// Convert votes to payload format with user names
	votes := make([]VoteInfo, 0, len(result.Votes))
	for userID, voteValue := range result.Votes {
		// Find user name from state
		userName := ""
		for _, user := range state.Users {
			if user.UserID == userID {
				userName = user.Name
				break
			}
		}
		votes = append(votes, VoteInfo{
			UserID:   userID,
			Value:    voteValue,
			Nickname: userName,
		})
	}

	// Broadcast votes revealed
	h.hub.BroadcastToRoom(client.RoomID, Message{
		Type: EventTypeVotesRevealed,
		Payload: VotesRevealedPayload{
			Votes:   votes,
			Average: result.Average,
		},
	}, nil)

	return nil
}

// handleClear processes a clear votes request
func (h *Handler) handleClear(ctx context.Context, client *Client) error {
	// Clear votes
	if err := h.votingService.ClearVotes(ctx, client.RoomID); err != nil {
		return fmt.Errorf("failed to clear votes: %w", err)
	}

	// Get updated room state to sync user voting status
	roomState, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	// Broadcast updated room state (includes reset user voting status)
	h.hub.BroadcastToRoom(client.RoomID, Message{
		Type:    EventTypeRoomState,
		Payload: h.convertRoomStateToPayload(roomState),
	}, nil)

	return nil
}

// handleUpdateNickname processes a nickname update request
func (h *Handler) handleUpdateNickname(ctx context.Context, client *Client, msg Message) error {
	var payload UpdateNicknamePayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid nickname payload: %w", err)
	}

	// Update nickname
	if err := h.userService.UpdateUserName(ctx, client.RoomID, client.UserID, payload.Nickname); err != nil {
		return fmt.Errorf("failed to update nickname: %w", err)
	}

	// Broadcast user updated
	h.hub.BroadcastToRoom(client.RoomID, Message{
		Type: EventTypeUserUpdated,
		Payload: UserUpdatedPayload{
			UserID:   client.UserID,
			Nickname: payload.Nickname,
		},
	}, nil)

	return nil
}

// handleSetTask processes a task description update
func (h *Handler) handleSetTask(ctx context.Context, client *Client, msg Message) error {
	var payload SetTaskPayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid task payload: %w", err)
	}

	// Update task description in state
	if err := h.roomService.UpdateTaskDescription(ctx, client.RoomID, payload.Description); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Send updated room state to all clients
	roomState, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, Message{
		Type:    EventTypeRoomState,
		Payload: h.convertRoomStateToPayload(roomState),
	}, nil)

	return nil
}

// sendRoomState sends the current room state to a client
func (h *Handler) sendRoomState(client *Client) error {
	ctx := context.Background()

	roomState, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	payload := h.convertRoomStateToPayload(roomState)

	data, err := json.Marshal(Message{
		Type:    EventTypeRoomState,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal room state: %w", err)
	}

	client.send <- data
	return nil
}

// convertRoomStateToPayload converts DTO to WebSocket payload
func (h *Handler) convertRoomStateToPayload(state *dto.RoomStateResp) RoomStatePayload {
	users := make([]UserPayload, len(state.Users))
	for i, user := range state.Users {
		users[i] = UserPayload{
			ID:       user.UserID,
			Name:     user.Name,
			IsVoted:  user.IsVoted,
			IsOnline: user.IsOnline,
		}
	}

	votes := make([]VoteInfo, len(state.Votes))
	for i, vote := range state.Votes {
		votes[i] = VoteInfo{
			UserID:   vote.UserID,
			Value:    vote.Value,
			Nickname: vote.UserName,
		}
	}

	return RoomStatePayload{
		RoomID:          state.RoomID,
		RoomName:        state.RoomName,
		Users:           users,
		Votes:           votes,
		IsRevealed:      state.IsRevealed,
		TaskDescription: state.TaskDescription,
		Average:         state.Average,
	}
}

// unmarshalPayload unmarshals a payload to the target type
func unmarshalPayload(payload interface{}, target interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
