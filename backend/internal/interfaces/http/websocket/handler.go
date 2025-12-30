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
	taskService   *application.TaskService
}

func NewHandler(
	hub *WsHub,
	roomService *application.RoomService,
	userService *application.UserService,
	votingService *application.VotingService,
	taskService *application.TaskService,
) *WsHandler {
	return &WsHandler{
		hub:           hub,
		roomService:   roomService,
		userService:   userService,
		votingService: votingService,
		taskService:   taskService,
	}
}

func (h *WsHandler) HandleConnection(c *fiber.Ctx) error {
	// IMPORTANT: Copy strings immediately to avoid fasthttp buffer reuse issues
	roomID := c.Params("id")
	roomIDCopy := string([]byte(roomID))

	userID := c.Query("userId")
	userIDCopy := string([]byte(userID))

	nickname := c.Query("nickname")
	nicknameCopy := string([]byte(nickname))

	log.Printf("[DEBUG] WebSocket connection - roomID: '%s', userID: '%s', path: '%s'", roomIDCopy, userIDCopy, c.Path())

	if roomIDCopy == "" || userIDCopy == "" || nicknameCopy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "roomId, userId, and nickname are required",
		})
	}

	// Upgrade the http conn to a WebSocket
	err := upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
		h.handleWebSocket(conn, roomIDCopy, userIDCopy, nicknameCopy)
	})

	return err
}

// manages the WebSocket lifecycle for a client
func (h *WsHandler) handleWebSocket(conn *websocket.Conn, roomID, userID, nickname string) {
	ctx := context.Background()

	// Remove user if they already exist (handles reconnection/stale connections)
	_ = h.userService.LeaveRoom(ctx, roomID, userID)

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

	// Send initial task list
	if err := h.sendTaskListSync(client); err != nil {
		log.Printf("Failed to send task list to user %s in room %s: %v", userID, roomID, err)
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

	case EventTypeCreateTask:
		return h.handleCreateTask(ctx, client, msg)

	case EventTypeUpdateTask:
		return h.handleUpdateTask(ctx, client, msg)

	case EventTypeDeleteTask:
		return h.handleDeleteTask(ctx, client, msg)

	case EventTypeReorderTasks:
		return h.handleReorderTasks(ctx, client, msg)

	case EventTypeSetActiveTask:
		return h.handleSetActiveTask(ctx, client, msg)

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

	// If votes are already revealed, recalculate and broadcast updated results
	if roomState.IsRevealed {
		result, err := h.votingService.RevealVotes(ctx, client.RoomID)
		if err != nil {
			return fmt.Errorf("failed to recalculate votes: %w", err)
		}

		votes := make([]VoteInfo, 0, len(result.Votes))
		for userID, voteValue := range result.Votes {
			userName := ""
			for _, user := range roomState.Users {
				if user.UserID == userID {
					userName = user.Name
					break
				}
			}
			votes = append(votes, VoteInfo{
				UserID:   userID,
				Value:    voteValue,
				UserName: userName,
			})
		}

		h.hub.BroadcastToRoom(client.RoomID, WsMessage{
			Type: EventTypeVotesRevealed,
			Payload: VotesRevealedPayload{
				Votes:   votes,
				Average: result.Average,
			},
		}, nil)
	}

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
			UserName: userName,
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
	// Get current state to check for votes
	state, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	// Save estimation to active task if votes were cast and revealed
	taskUpdated := false
	if state.IsRevealed && len(state.Votes) > 0 {
		// Calculate result to determine estimation
		result, err := h.votingService.RevealVotes(ctx, client.RoomID)
		if err != nil {
			log.Printf("Warning: failed to get vote result: %v", err)
		} else {
			estimation := h.determineEstimation(result)

			// Get the active task ID from room state
			activeTaskID, err := h.roomService.GetActiveTask(client.RoomID)
			if err != nil {
				log.Printf("Warning: failed to get active task: %v", err)
			}

			// Save estimation to the active task if set, otherwise fallback to next unestimated
			if activeTaskID != "" {
				if err := h.taskService.SaveEstimationToTask(ctx, activeTaskID, estimation); err != nil {
					log.Printf("Warning: failed to save estimation to active task: %v", err)
				} else {
					taskUpdated = true
				}
			} else {
				if err := h.taskService.SaveEstimation(ctx, client.RoomID, estimation); err != nil {
					log.Printf("Warning: failed to save estimation: %v", err)
				} else {
					taskUpdated = true
				}
			}
		}
	}

	// Clear votes
	if err := h.votingService.ClearVotes(ctx, client.RoomID); err != nil {
		return fmt.Errorf("failed to clear votes: %w", err)
	}

	// Move to next unestimated task
	nextTask, err := h.taskService.GetNextUnestimatedTask(ctx, client.RoomID)
	if err != nil {
		log.Printf("Warning: failed to get next unestimated task: %v", err)
	}

	// Get updated room state
	newRoomState, err := h.roomService.GetRoomState(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room state: %w", err)
	}

	// Get updated task list to reflect saved estimations
	tasks, err := h.taskService.GetRoomTasks(ctx, client.RoomID)
	if err != nil {
		log.Printf("Warning: failed to get tasks after clear: %v", err)
	}

	// Broadcast votes cleared event
	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type:    EventTypeVotesCleared,
		Payload: VotesClearedPayload{},
	}, nil)

	// Broadcast room state
	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type:    EventTypeRoomState,
		Payload: h.convertRoomStateToPayload(newRoomState),
	}, nil)

	// Broadcast updated task list to show saved estimations
	if taskUpdated || tasks != nil {
		if tasks == nil {
			tasks, err = h.taskService.GetRoomTasks(ctx, client.RoomID)
			if err != nil {
				log.Printf("Warning: failed to get tasks for broadcast: %v", err)
			}
		}
		if tasks != nil {
			taskPayloads := make([]TaskPayload, len(tasks))
			for i, task := range tasks {
				taskPayloads[i] = convertTaskToPayload(task)
			}
			h.hub.BroadcastToRoom(client.RoomID, WsMessage{
				Type: EventTypeTaskListSync,
				Payload: TaskListSyncPayload{
					Tasks: taskPayloads,
				},
			}, nil)
		}
	}

	// If there's a next task, broadcast that too
	if nextTask != nil {
		h.hub.BroadcastToRoom(client.RoomID, WsMessage{
			Type: EventTypeActiveTaskSet,
			Payload: SetActiveTaskPayload{
				TaskID: nextTask.ID,
			},
		}, nil)
	}

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
			UserName: vote.UserName,
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

// Task-related handlers
func (h *WsHandler) handleCreateTask(ctx context.Context, client *Client, msg WsMessage) error {
	var payload CreateTaskPayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid create task payload: %w", err)
	}

	req := &dto.CreateTaskReq{
		Headline:    payload.Headline,
		Description: payload.Description,
		TrackerLink: payload.TrackerLink,
	}

	task, err := h.taskService.CreateTask(ctx, client.RoomID, req)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type:    EventTypeTaskCreated,
		Payload: convertTaskToPayload(task),
	}, nil)

	return nil
}

func (h *WsHandler) handleUpdateTask(ctx context.Context, client *Client, msg WsMessage) error {
	var payload UpdateTaskPayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid update task payload: %w", err)
	}

	req := &dto.UpdateTaskReq{
		Headline:    payload.Headline,
		Description: payload.Description,
		TrackerLink: payload.TrackerLink,
	}

	task, err := h.taskService.UpdateTask(ctx, payload.TaskID, req)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type:    EventTypeTaskUpdated,
		Payload: convertTaskToPayload(task),
	}, nil)

	return nil
}

func (h *WsHandler) handleDeleteTask(ctx context.Context, client *Client, msg WsMessage) error {
	var payload DeleteTaskPayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid delete task payload: %w", err)
	}

	if err := h.taskService.DeleteTask(ctx, payload.TaskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type: EventTypeTaskDeleted,
		Payload: fiber.Map{
			"taskId": payload.TaskID,
		},
	}, nil)

	return nil
}

func (h *WsHandler) handleReorderTasks(ctx context.Context, client *Client, msg WsMessage) error {
	var payload ReorderTasksPayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid reorder payload: %w", err)
	}

	req := &dto.ReorderTasksReq{
		TaskIDs: payload.TaskIDs,
	}

	if err := h.taskService.ReorderTasks(ctx, client.RoomID, req); err != nil {
		return fmt.Errorf("failed to reorder tasks: %w", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type:    EventTypeTasksReordered,
		Payload: payload,
	}, nil)

	return nil
}

func (h *WsHandler) handleSetActiveTask(ctx context.Context, client *Client, msg WsMessage) error {
	var payload SetActiveTaskPayload
	if err := unmarshalPayload(msg.Payload, &payload); err != nil {
		return fmt.Errorf("invalid set active task payload: %w", err)
	}

	// Set the active task in room state
	if err := h.roomService.SetActiveTask(client.RoomID, payload.TaskID); err != nil {
		log.Printf("Warning: failed to set active task: %v", err)
	}

	h.hub.BroadcastToRoom(client.RoomID, WsMessage{
		Type:    EventTypeActiveTaskSet,
		Payload: payload,
	}, nil)

	return nil
}

func (h *WsHandler) sendTaskListSync(client *Client) error {
	ctx := context.Background()

	tasks, err := h.taskService.GetRoomTasks(ctx, client.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	taskPayloads := make([]TaskPayload, len(tasks))
	for i, task := range tasks {
		taskPayloads[i] = convertTaskToPayload(task)
	}

	payload := TaskListSyncPayload{
		Tasks: taskPayloads,
	}

	data, err := json.Marshal(WsMessage{
		Type:    EventTypeTaskListSync,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal task list: %w", err)
	}

	client.send <- data
	return nil
}

func convertTaskToPayload(task *dto.TaskResp) TaskPayload {
	return TaskPayload{
		ID:          task.ID,
		RoomID:      task.RoomID,
		Headline:    task.Headline,
		Description: task.Description,
		TrackerLink: task.TrackerLink,
		Estimation:  task.Estimation,
		Position:    task.Position,
	}
}

// determineEstimation calculates the estimation value from reveal results
// Uses average for numeric votes, or consensus/most common vote for non-numeric
func (h *WsHandler) determineEstimation(result *dto.RevealVotesResp) string {
	// If we have a numeric average, use it
	if result.Average != nil {
		return fmt.Sprintf("%.1f", *result.Average)
	}

	// No numeric average - find consensus or most common vote
	if len(result.Votes) == 0 {
		return "?"
	}

	// Count vote occurrences
	voteCounts := make(map[string]int)
	for _, voteValue := range result.Votes {
		voteCounts[voteValue]++
	}

	// Find the most common vote
	var maxCount int
	var mostCommon string
	for value, count := range voteCounts {
		if count > maxCount {
			maxCount = count
			mostCommon = value
		}
	}

	// If everyone voted the same, use that value
	if maxCount == len(result.Votes) {
		return mostCommon
	}

	// No clear consensus - mark as uncertain
	return "?"
}
