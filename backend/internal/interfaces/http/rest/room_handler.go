package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vitaly-stepin/agile_party/internal/application"
	"github.com/vitaly-stepin/agile_party/internal/application/dto"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

// RoomHandler handles HTTP requests for room operations
type RoomHandler struct {
	roomService   *application.RoomService
	userService   *application.UserService
	votingService *application.VotingService
}

// NewRoomHandler creates a new RoomHandler
func NewRoomHandler(
	roomService *application.RoomService,
	userService *application.UserService,
	votingService *application.VotingService,
) *RoomHandler {
	return &RoomHandler{
		roomService:   roomService,
		userService:   userService,
		votingService: votingService,
	}
}

// CreateRoom handles POST /api/rooms
func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	var req dto.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	resp, err := h.roomService.CreateRoom(c.Context(), &req)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetRoom handles GET /api/rooms/:id
func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	resp, err := h.roomService.GetRoom(c.Context(), roomID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(resp)
}

// GetRoomState handles GET /api/rooms/:id/state
func (h *RoomHandler) GetRoomState(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	resp, err := h.roomService.GetRoomState(c.Context(), roomID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(resp)
}

// JoinRoom handles POST /api/rooms/:id/users
func (h *RoomHandler) JoinRoom(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	var req struct {
		UserID   string `json:"userId"`
		UserName string `json:"userName"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" || req.UserName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId and userName are required",
		})
	}

	user, err := h.userService.JoinRoom(c.Context(), roomID, req.UserID, req.UserName)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// LeaveRoom handles DELETE /api/rooms/:id/users/:userId
func (h *RoomHandler) LeaveRoom(c *fiber.Ctx) error {
	roomID := c.Params("id")
	userID := c.Params("userId")

	if roomID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID and User ID are required",
		})
	}

	if err := h.userService.LeaveRoom(c.Context(), roomID, userID); err != nil {
		return handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateUserName handles PATCH /api/rooms/:id/users/:userId
func (h *RoomHandler) UpdateUserName(c *fiber.Ctx) error {
	roomID := c.Params("id")
	userID := c.Params("userId")

	if roomID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID and User ID are required",
		})
	}

	var req struct {
		UserName string `json:"userName"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userName is required",
		})
	}

	user, err := h.userService.UpdateUserName(c.Context(), roomID, userID, req.UserName)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(user)
}

// SubmitVote handles POST /api/rooms/:id/votes
func (h *RoomHandler) SubmitVote(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	var req dto.SubmitVoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" || req.Value == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId and value are required",
		})
	}

	if err := h.votingService.SubmitVote(c.Context(), roomID, req.UserID, req.Value); err != nil {
		return handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RevealVotes handles POST /api/rooms/:id/reveal
func (h *RoomHandler) RevealVotes(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	resp, err := h.votingService.RevealVotes(c.Context(), roomID)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(resp)
}

// ClearVotes handles POST /api/rooms/:id/clear
func (h *RoomHandler) ClearVotes(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	if err := h.votingService.ClearVotes(c.Context(), roomID); err != nil {
		return handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HealthCheck handles GET /api/health
func (h *RoomHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// handleError maps domain errors to HTTP status codes
func handleError(c *fiber.Ctx, err error) error {
	switch err {
	case room.ErrRoomNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	case room.ErrUserNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	case room.ErrInvalidRoomName, room.ErrInvalidUserName, room.ErrInvalidVote:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	case room.ErrVotesAlreadyRevealed, room.ErrVotesNotRevealed:
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}
}
