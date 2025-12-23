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
		// Allow all origins for MVP (configure properly in prod)
		return true
	},
}

type WsHandler struct {
	hub           *WsHub
	roomService   *application.RoomService
	userService   *application.UserService
	votingService *application.VotingService
}

func NewHandler(
	hub *WsHub,
	roomService *application.RoomService,
	userService *application.UserService,
	votingService *application.VotingService,
) *WsHandler {
	return &WsHandler{
		hub:           hub,
		roomService:   roomService,
		userService:   userService,
		votingService: votingService,
	}
}

func (h *WsHandler) HandleConnection(c *fiber.Ctx) error {
	roomID := c.Params("id")
	userID := c.Query("userId")
	nickname := c.Query("nickname")

	log.Printf("[DEBUG] WebSocket connection - roomID: '%s', userID: '%s', path: '%s'", roomID, userID, c.Path())

	if roomID == "" || userID == "" || nickname == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "roomId, userId, and nickname are required",
		})
	}

	// Upgrade the http conn to a WebSocket
	err := upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
		h.handleWebSocket(conn, roomID, userID, nickname)
	})

	return err
}

// manages the WebSocket lifecycle for a client
func (h *WsHandler) handleWebSocket(conn *websocket.Conn, roomID, userID, nickname string) {
	ctx := context.Background()

	if err := h.userService.JoinRoom(ctx, roomID, userID, nickname); err != nil {
		log.Printf("Failed to join room %s for user %s: %v", roomID, userID, err)
		conn.WriteJSON(WsMessage{
			Type: EventTypeError,
			Payload: ErrorPayload{
				Message: "Failed to join room: " + err.Error(),
				Code:    "JOIN_FAILED",
			},
		})
		conn.Close()
		return
	}

	log.Printf("[DEBUG] Creating client - roomID: '%s', userID: '%s'", roomID, userID)
	client := NewClient(conn, h.hub, roomID, userID, h)
	log.Printf("[DEBUG] Client created - client.RoomID: '%s', client.UserID: '%s'", client.RoomID, client.UserID)
	h.hub.register <- client

	if err := h.sendRoomState(client); err != nil {
		log.Printf("Failed to send initial state to user %s in room %s: %v", userID, roomID, err)
	}

	h.hub.BroadcastToRoom(roomID, WsMessage{
		Type: EventTypeUserJoined,
		Payload: UserJoinedPayload{
			UserID:   userID,
			Nickname: nickname,
		},
	}, client)

	defer func() {
		if err := h.userService.LeaveRoom(ctx, roomID, userID); err != nil {
			log.Printf("Failed to remove user %s from room %s: %v", userID, roomID, err)
		}

		h.hub.BroadcastToRoom(roomID, WsMessage{
			Type: EventTypeUserLeft,
			Payload: UserLeftPayload{
				UserID: userID,
			},
		}, nil)

		log.Printf("User %s disconnected from room %s", userID, roomID)
	}()

	client.Start()
}

func (h *WsHandler) HandleMessage(client *Client, msg WsMessage) error {
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

func (h *WsHandler) handleVote(ctx context.Context, client *Client, msg WsMessage) error {
	var payload VotePayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid vote payload: %w", err)
	}

	if err := h.votingService.SubmitVote(ctx, client.RoomID, client.UserID, payload.Value); err != nil {
		return fmt.Errorf("failed to submit vote: %w", err)
	}

	roomState, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type:    EventTypeRoomState,
		Payload: h.convertRoomStateToPayload(roomState),
	}, nil)

	return nil
}

func (h *WsHandler) handleReveal(ctx context.Context, client *Client) error {
	result, err := h.votingService.RevealVotes(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to reveal votes: %w", err)
	}

	state, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	votes := make([]VoteInfo, 0, len(result.Votes))
	for userID, voteValue := range result.Votes {
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

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type: EventTypeVotesRevealed,
		Payload: VotesRevealedPayload{
			Votes:   votes,
			Average: result.Average,
		},
	}, nil)

	return nil
}

func (h *WsHandler) handleClear(ctx context.Context, client *Client) error {
	if err := h.votingService.ClearVotes(ctx, client.RoomID); err != nil {
		return fmt.Errorf("failed to clear votes: %w", err)
	}

	roomState, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type:    EventTypeRoomState,
		Payload: h.convertRoomStateToPayload(roomState),
	}, nil)

	return nil
}

func (h *WsHandler) handleUpdateNickname(ctx context.Context, client *Client, msg WsMessage) error {
	var payload UpdateNicknamePayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid nickname payload: %w", err)
	}

	if err := h.userService.UpdateUserName(ctx, client.RoomID, client.UserID, payload.Nickname); err != nil {
		return fmt.Errorf("failed to update nickname: %w", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type: EventTypeUserUpdated,
		Payload: UserUpdatedPayload{
			UserID:   client.UserID,
			Nickname: payload.Nickname,
		},
	}, nil)

	return nil
}

func (h *WsHandler) handleSetTask(ctx context.Context, client *Client, msg WsMessage) error {
	var payload SetTaskPayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid task payload: %w", err)
	}

	if err := h.roomService.UpdateTaskDescription(ctx, client.RoomID, payload.Description); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	roomState, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type:    EventTypeRoomState,
		Payload: h.convertRoomStateToPayload(roomState),
	}, nil)

	return nil
}

func (h *WsHandler) sendRoomState(client *Client) error {
	ctx := context.Background()

	roomState, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	payload := h.convertRoomStateToPayload(roomState)

	data, err := json.Marshal(WsMessage{
		Type:    EventTypeRoomState,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal room state: %w", err)
	}

	client.send <- data
	return nil
}

func (h *WsHandler) convertRoomStateToPayload(state *dto.RoomStateResp) RoomStatePayload {
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

func unmarshalPayload(payload interface{}, target interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
