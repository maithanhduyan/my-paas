package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/config"
	"github.com/my-paas/server/crypto"
	"github.com/my-paas/server/store"
)

func AuthRequired(s *store.Store, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If no users exist, skip auth (setup mode)
		count, _ := s.GetUserCount()
		if count == 0 {
			return c.Next()
		}

		// Maintenance mode check
		if cfg.MaintenanceMode {
			return c.Status(503).JSON(fiber.Map{"error": "service under maintenance"})
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

		// Try JWT token first
		if claims, err := crypto.ValidateJWT(token, cfg.Secret); err == nil {
			c.Locals("user_id", claims.Sub)
			c.Locals("user_role", claims.Role)
			c.Locals("username", claims.Username)
			c.Locals("auth_type", "jwt")
			return c.Next()
		}

		// Try API key (prefix: mpk_)
		if strings.HasPrefix(token, "mpk_") {
			keyHash := crypto.HashAPIKey(token)
			apiKey, err := s.GetAPIKeyByHash(keyHash)
			if err != nil || apiKey == nil {
				return c.Status(401).JSON(fiber.Map{"error": "invalid API key"})
			}
			if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
				return c.Status(401).JSON(fiber.Map{"error": "API key expired"})
			}
			// Update last used
			s.UpdateAPIKeyLastUsed(apiKey.ID)

			user, err := s.GetUser(apiKey.UserID)
			if err != nil || user == nil {
				return c.Status(401).JSON(fiber.Map{"error": "user not found"})
			}

			c.Locals("user_id", user.ID)
			c.Locals("user_role", user.Role)
			c.Locals("username", user.Username)
			c.Locals("auth_type", "api_key")
			c.Locals("api_key_scopes", apiKey.Scopes)
			return c.Next()
		}

		// Fallback: session token (backward compatible)
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
		c.Locals("auth_type", "session")
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

// ScopeRequired checks API key scopes for the given permission.
func ScopeRequired(scope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authType, _ := c.Locals("auth_type").(string)
		if authType != "api_key" {
			return c.Next() // JWT and session tokens have full access
		}

		scopes, _ := c.Locals("api_key_scopes").(string)
		if scopes == "*" {
			return c.Next()
		}

		for _, s := range strings.Split(scopes, ",") {
			if strings.TrimSpace(s) == scope {
				return c.Next()
			}
		}
		return c.Status(403).JSON(fiber.Map{"error": "insufficient API key scope: " + scope})
	}
}
