package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vitaly-stepin/agile_party/internal/application"
	"github.com/vitaly-stepin/agile_party/internal/application/dto"
)

type RoomHandler struct {
	roomService   *application.RoomService
	userService   *application.UserService
	votingService *application.VotingService
}

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

func (h *RoomHandler) NewRoom(c *fiber.Ctx) error {
	var req dto.NewRoomReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.VotingSystem == "" {
		req.VotingSystem = "fibonacci"
	}

	response, err := h.roomService.NewRoom(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	response, err := h.roomService.GetRoom(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

func (h *RoomHandler) GetRoomState(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	response, err := h.roomService.GetRoomState(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

func (h *RoomHandler) JoinRoom(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	var req struct {
		UserID   string `json:"user_id"`
		UserName string `json:"user_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" || req.UserName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id and user_name are required",
		})
	}

	if err := h.userService.JoinRoom(c.Context(), roomID, req.UserID, req.UserName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User joined successfully",
	})
}

func (h *RoomHandler) LeaveRoom(c *fiber.Ctx) error {
	roomID := c.Params("id")
	userID := c.Params("userId")

	if roomID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID and User ID are required",
		})
	}

	if err := h.userService.LeaveRoom(c.Context(), roomID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User left successfully",
	})
}

func (h *RoomHandler) UpdateUserName(c *fiber.Ctx) error {
	roomID := c.Params("id")
	userID := c.Params("userId")

	if roomID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID and User ID are required",
		})
	}

	var req struct {
		UserName string `json:"user_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_name is required",
		})
	}

	if err := h.userService.UpdateUserName(c.Context(), roomID, userID, req.UserName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User name updated successfully",
	})
}

func (h *RoomHandler) SubmitVote(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	var req dto.SubmitVoteReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" || req.Value == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id and value are required",
		})
	}

	if err := h.votingService.SubmitVote(c.Context(), roomID, req.UserID, req.Value); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Vote submitted successfully",
	})
}

func (h *RoomHandler) RevealVotes(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	response, err := h.votingService.RevealVotes(c.Context(), roomID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

func (h *RoomHandler) ClearVotes(c *fiber.Ctx) error {
	roomID := c.Params("id")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room ID is required",
		})
	}

	if err := h.votingService.ClearVotes(c.Context(), roomID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Votes cleared successfully",
	})
}

func (h *RoomHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Agile Party API is running",
	})
}
