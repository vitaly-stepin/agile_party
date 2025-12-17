package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/vitaly-stepin/agile_party/internal/application"
	"github.com/vitaly-stepin/agile_party/internal/application/dto"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a WebSocket connection
type Client struct {
	hub *Hub

	// WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Room ID this client is connected to
	roomID string

	// User ID
	userID string

	// User name
	userName string

	// Application services
	userService   *application.UserService
	votingService *application.VotingService
}

// NewClient creates a new WebSocket client
func NewClient(
	hub *Hub,
	conn *websocket.Conn,
	roomID, userID, userName string,
	userService *application.UserService,
	votingService *application.VotingService,
) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		roomID:        roomID,
		userID:        userID,
		userName:      userName,
		userService:   userService,
		votingService: votingService,
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()

		// Remove user from room when disconnected
		ctx := context.Background()
		if err := c.userService.LeaveRoom(ctx, c.roomID, c.userID); err != nil {
			log.Printf("Error removing user from room: %v", err)
		}

		// Broadcast user left event
		c.broadcastUserLeft()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
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
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(message []byte) {
	var event Event
	if err := json.Unmarshal(message, &event); err != nil {
		c.sendError("Invalid message format", "INVALID_FORMAT")
		return
	}

	ctx := context.Background()

	switch event.Type {
	case EventTypeVote:
		c.handleVote(ctx, event.Payload)

	case EventTypeReveal:
		c.handleReveal(ctx)

	case EventTypeClear:
		c.handleClear(ctx)

	case EventTypeUpdateNickname:
		c.handleUpdateNickname(ctx, event.Payload)

	default:
		c.sendError("Unknown event type", "UNKNOWN_EVENT")
	}
}

// handleVote handles vote submission
func (c *Client) handleVote(ctx context.Context, payload interface{}) {
	data, ok := payload.(map[string]interface{})
	if !ok {
		c.sendError("Invalid vote payload", "INVALID_PAYLOAD")
		return
	}

	value, ok := data["value"].(string)
	if !ok {
		c.sendError("Vote value must be a string", "INVALID_VOTE")
		return
	}

	if err := c.votingService.SubmitVote(ctx, c.roomID, c.userID, value); err != nil {
		c.sendError(err.Error(), "VOTE_ERROR")
		return
	}

	// Broadcast vote submitted event
	c.broadcastVoteSubmitted()
}

// handleReveal handles vote reveal
func (c *Client) handleReveal(ctx context.Context) {
	resp, err := c.votingService.RevealVotes(ctx, c.roomID)
	if err != nil {
		c.sendError(err.Error(), "REVEAL_ERROR")
		return
	}

	// Broadcast votes revealed event
	c.broadcastVotesRevealed(resp)
}

// handleClear handles vote clearing
func (c *Client) handleClear(ctx context.Context) {
	if err := c.votingService.ClearVotes(ctx, c.roomID); err != nil {
		c.sendError(err.Error(), "CLEAR_ERROR")
		return
	}

	// Broadcast votes cleared event
	c.broadcastVotesCleared()
}

// handleUpdateNickname handles nickname update
func (c *Client) handleUpdateNickname(ctx context.Context, payload interface{}) {
	data, ok := payload.(map[string]interface{})
	if !ok {
		c.sendError("Invalid nickname payload", "INVALID_PAYLOAD")
		return
	}

	nickname, ok := data["nickname"].(string)
	if !ok {
		c.sendError("Nickname must be a string", "INVALID_NICKNAME")
		return
	}

	user, err := c.userService.UpdateUserName(ctx, c.roomID, c.userID, nickname)
	if err != nil {
		c.sendError(err.Error(), "UPDATE_ERROR")
		return
	}

	c.userName = user.Name

	// Broadcast user updated event
	c.broadcastUserUpdated()
}

// sendError sends an error event to the client
func (c *Client) sendError(message, code string) {
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
		return
	}

	select {
	case c.send <- data:
	default:
		log.Printf("Client send channel full, dropping error message")
	}
}

// broadcastVoteSubmitted broadcasts vote submitted event to all clients in the room
func (c *Client) broadcastVoteSubmitted() {
	event := Event{
		Type: EventTypeVoteSubmitted,
		Payload: VoteSubmittedPayload{
			UserID:  c.userID,
			IsVoted: true,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling vote submitted event: %v", err)
		return
	}

	c.hub.BroadcastToRoom(c.roomID, data)
}

// broadcastVotesRevealed broadcasts votes revealed event to all clients in the room
func (c *Client) broadcastVotesRevealed(resp *dto.RevealVotesResponse) {
	votes := make([]VotePayload2, 0, len(resp.Votes))
	for userID, value := range resp.Votes {
		votes = append(votes, VotePayload2{
			UserID: userID,
			Value:  value,
		})
	}

	event := Event{
		Type: EventTypeVotesRevealed,
		Payload: VotesRevealedPayload{
			Votes:   votes,
			Average: resp.Average,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling votes revealed event: %v", err)
		return
	}

	c.hub.BroadcastToRoom(c.roomID, data)
}

// broadcastVotesCleared broadcasts votes cleared event to all clients in the room
func (c *Client) broadcastVotesCleared() {
	event := Event{
		Type:    EventTypeVotesCleared,
		Payload: struct{}{},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling votes cleared event: %v", err)
		return
	}

	c.hub.BroadcastToRoom(c.roomID, data)
}

// broadcastUserUpdated broadcasts user updated event to all clients in the room
func (c *Client) broadcastUserUpdated() {
	event := Event{
		Type: EventTypeUserUpdated,
		Payload: UserUpdatedPayload{
			UserID: c.userID,
			Name:   c.userName,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling user updated event: %v", err)
		return
	}

	c.hub.BroadcastToRoom(c.roomID, data)
}

// broadcastUserLeft broadcasts user left event to all clients in the room
func (c *Client) broadcastUserLeft() {
	event := Event{
		Type: EventTypeUserLeft,
		Payload: UserLeftPayload{
			UserID: c.userID,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling user left event: %v", err)
		return
	}

	c.hub.BroadcastToRoom(c.roomID, data)
}
