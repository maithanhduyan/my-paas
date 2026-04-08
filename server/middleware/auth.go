package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/store"
)

func AuthRequired(s *store.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If no users exist, skip auth (setup mode)
		count, _ := s.GetUserCount()
		if count == 0 {
			return c.Next()
		}

		auth := c.Get("Authorization")
		token := ""
		if auth != "" {
			token = strings.TrimPrefix(auth, "Bearer ")
			if token == auth {
				token = ""
			}
		}
		// Also accept token as query param (for SSE streams)
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}

		session, err := s.GetSessionByToken(token)
		if err != nil || session == nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		if time.Now().After(session.ExpiresAt) {
			s.DeleteSessionByToken(token)
			return c.Status(401).JSON(fiber.Map{"error": "token expired"})
		}

		// Load user to get role and username
		user, err := s.GetUser(session.UserID)
		if err != nil || user == nil {
			return c.Status(401).JSON(fiber.Map{"error": "user not found"})
		}

		c.Locals("user_id", session.UserID)
		c.Locals("user_role", user.Role)
		c.Locals("username", user.Username)
		return c.Next()
	}
}

// RoleRequired restricts access to users with one of the specified roles.
func RoleRequired(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("user_role").(string)
		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}
		return c.Status(403).JSON(fiber.Map{"error": "forbidden: insufficient permissions"})
	}
}
