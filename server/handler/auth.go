package handler

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/my-paas/server/crypto"
	"github.com/my-paas/server/model"
)

func (h *Handler) AuthStatus(c *fiber.Ctx) error {
	count, _ := h.Store.GetUserCount()

	auth := c.Get("Authorization")
	if auth != "" {
		token := strings.TrimPrefix(auth, "Bearer ")
		if token != auth {
			session, err := h.Store.GetSessionByToken(token)
			if err == nil && session != nil {
				user, _ := h.Store.GetUser(session.UserID)
				return c.JSON(model.AuthStatus{
					Authenticated: true,
					User:          user,
				})
			}
		}
	}

	return c.JSON(model.AuthStatus{
		SetupRequired: count == 0,
	})
}

func (h *Handler) Setup(c *fiber.Ctx) error {
	count, _ := h.Store.GetUserCount()
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "setup already completed"})
	}

	var input model.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Username == "" || input.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username and password required"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal error"})
	}

	user, err := h.Store.CreateUser(input.Username, string(hash), "admin")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	token := generateToken()
	session, err := h.Store.CreateSession(user.ID, token)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	response := fiber.Map{
		"user":    user,
		"token":   session.Token,
		"expires": session.ExpiresAt,
	}

	// Include JWT tokens if enterprise config is available
	if h.Config != nil && h.Config.Secret != "" {
		accessToken, _ := crypto.GenerateJWT(user.ID, user.Username, user.Role, "access", h.Config.Secret, h.Config.JWTExpiry)
		refreshToken, _ := crypto.GenerateJWT(user.ID, user.Username, user.Role, "refresh", h.Config.Secret, h.Config.RefreshExpiry)
		if accessToken != "" {
			response["access_token"] = accessToken
			response["refresh_token"] = refreshToken
			response["expires_in"] = int(h.Config.JWTExpiry.Seconds())
		}
	}

	return c.Status(201).JSON(response)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var input model.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := h.Store.GetUserByUsername(input.Username)
	if err != nil || user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token := generateToken()
	session, err := h.Store.CreateSession(user.ID, token)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	response := fiber.Map{
		"user":    user,
		"token":   session.Token,
		"expires": session.ExpiresAt,
	}

	// Include JWT tokens if enterprise config is available
	if h.Config != nil && h.Config.Secret != "" {
		accessToken, _ := crypto.GenerateJWT(user.ID, user.Username, user.Role, "access", h.Config.Secret, h.Config.JWTExpiry)
		refreshToken, _ := crypto.GenerateJWT(user.ID, user.Username, user.Role, "refresh", h.Config.Secret, h.Config.RefreshExpiry)
		if accessToken != "" {
			response["access_token"] = accessToken
			response["refresh_token"] = refreshToken
			response["expires_in"] = int(h.Config.JWTExpiry.Seconds())
		}
	}

	return c.JSON(response)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if auth != "" {
		token := strings.TrimPrefix(auth, "Bearer ")
		if token != auth {
			h.Store.DeleteSessionByToken(token)
		}
	}
	return c.JSON(fiber.Map{"message": "logged out"})
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
