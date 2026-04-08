package handler

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/my-paas/server/model"
	"github.com/my-paas/server/store"
)

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	users, err := h.Store.ListUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if users == nil {
		users = []model.User{}
	}
	return c.JSON(users)
}

func (h *Handler) UpdateUserRole(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)

	var input struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Role != "admin" && input.Role != "member" && input.Role != "viewer" {
		return c.Status(400).JSON(fiber.Map{"error": "role must be admin, member, or viewer"})
	}

	// Prevent self-demotion
	if id == userID {
		return c.Status(400).JSON(fiber.Map{"error": "cannot change your own role"})
	}

	if err := h.Store.UpdateUserRole(id, input.Role); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.Store.AddAuditLog(userID, username, "update", "user", id, "role changed to "+input.Role)
	return c.JSON(fiber.Map{"message": "role updated"})
}

func (h *Handler) DeleteUserAccount(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)

	if id == userID {
		return c.Status(400).JSON(fiber.Map{"error": "cannot delete your own account"})
	}

	if err := h.Store.DeleteUser(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.Store.AddAuditLog(userID, username, "delete", "user", id, "user deleted")
	return c.JSON(fiber.Map{"message": "user deleted"})
}

func (h *Handler) InviteUser(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)

	var input struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "email is required"})
	}
	if input.Role == "" {
		input.Role = "member"
	}
	if input.Role != "admin" && input.Role != "member" && input.Role != "viewer" {
		return c.Status(400).JSON(fiber.Map{"error": "role must be admin, member, or viewer"})
	}

	token := generateInviteToken()
	inv := &model.Invitation{
		ID:        store.NewID(),
		Email:     input.Email,
		Role:      input.Role,
		Token:     token,
		Used:      false,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.Store.CreateInvitation(inv); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.Store.AddAuditLog(userID, username, "create", "invitation", inv.ID, "invited "+input.Email+" as "+input.Role)
	return c.Status(201).JSON(inv)
}

func (h *Handler) ListInvitations(c *fiber.Ctx) error {
	invs, err := h.Store.ListInvitations()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if invs == nil {
		invs = []model.Invitation{}
	}
	return c.JSON(invs)
}

func (h *Handler) RegisterWithInvite(c *fiber.Ctx) error {
	var input struct {
		Token    string `json:"token"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Token == "" || input.Username == "" || input.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "token, username, and password required"})
	}

	inv, err := h.Store.GetInvitationByToken(input.Token)
	if err != nil || inv == nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid invitation token"})
	}
	if inv.Used {
		return c.Status(400).JSON(fiber.Map{"error": "invitation already used"})
	}
	if time.Now().After(inv.ExpiresAt) {
		return c.Status(400).JSON(fiber.Map{"error": "invitation expired"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal error"})
	}

	user, err := h.Store.CreateUser(input.Username, string(hash), inv.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.Store.MarkInvitationUsed(inv.ID)

	// Auto-login
	sessionToken := generateInviteToken()
	session, err := h.Store.CreateSession(user.ID, sessionToken)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.Store.AddAuditLog(user.ID, user.Username, "create", "user", user.ID, "registered via invitation")
	return c.Status(201).JSON(fiber.Map{
		"user":    user,
		"token":   session.Token,
		"expires": session.ExpiresAt,
	})
}

func (h *Handler) ListAuditLogs(c *fiber.Ctx) error {
	logs, err := h.Store.ListAuditLogs(200)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if logs == nil {
		logs = []model.AuditLog{}
	}
	return c.JSON(logs)
}

func generateInviteToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
