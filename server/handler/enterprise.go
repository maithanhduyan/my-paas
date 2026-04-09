package handler

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/crypto"
	"github.com/my-paas/server/model"
	"github.com/my-paas/server/store"
)

// --- Organizations ---

func (h *Handler) ListOrganizations(c *fiber.Ctx) error {
	orgs, err := h.Store.ListOrganizations()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orgs)
}

func (h *Handler) CreateOrganization(c *fiber.Ctx) error {
	var input model.CreateOrgInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	if input.Slug == "" {
		input.Slug = strings.ToLower(strings.ReplaceAll(input.Name, " ", "-"))
	}

	org, err := h.Store.CreateOrganization(input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Add creator as owner
	userID, _ := c.Locals("user_id").(string)
	if userID != "" {
		h.Store.AddOrgMember(org.ID, userID, "owner")
	}

	h.audit(c, "create", "organization", org.ID, "name: "+org.Name)
	return c.Status(201).JSON(org)
}

func (h *Handler) GetOrganization(c *fiber.Ctx) error {
	org, err := h.Store.GetOrganization(c.Params("id"))
	if err != nil || org == nil {
		return c.Status(404).JSON(fiber.Map{"error": "organization not found"})
	}
	return c.JSON(org)
}

func (h *Handler) UpdateOrganization(c *fiber.Ctx) error {
	var input model.UpdateOrgInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	org, err := h.Store.UpdateOrganization(c.Params("id"), input)
	if err != nil || org == nil {
		return c.Status(404).JSON(fiber.Map{"error": "organization not found"})
	}

	h.audit(c, "update", "organization", org.ID, "")
	return c.JSON(org)
}

func (h *Handler) DeleteOrganization(c *fiber.Ctx) error {
	if err := h.Store.DeleteOrganization(c.Params("id")); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	h.audit(c, "delete", "organization", c.Params("id"), "")
	return c.JSON(fiber.Map{"message": "organization deleted"})
}

func (h *Handler) ListOrgMembers(c *fiber.Ctx) error {
	members, err := h.Store.ListOrgMembers(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(members)
}

func (h *Handler) AddOrgMember(c *fiber.Ctx) error {
	var input model.AddOrgMemberInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.UserID == "" || input.Role == "" {
		return c.Status(400).JSON(fiber.Map{"error": "user_id and role required"})
	}
	if err := h.Store.AddOrgMember(c.Params("id"), input.UserID, input.Role); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	h.audit(c, "add_member", "organization", c.Params("id"), "user: "+input.UserID+" role: "+input.Role)
	return c.JSON(fiber.Map{"message": "member added"})
}

func (h *Handler) RemoveOrgMember(c *fiber.Ctx) error {
	if err := h.Store.RemoveOrgMember(c.Params("id"), c.Params("userId")); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	h.audit(c, "remove_member", "organization", c.Params("id"), "user: "+c.Params("userId"))
	return c.JSON(fiber.Map{"message": "member removed"})
}

// GetOrgQuotaUsage returns current resource usage vs quotas.
func (h *Handler) GetOrgQuotaUsage(c *fiber.Ctx) error {
	orgID := c.Params("id")
	org, err := h.Store.GetOrganization(orgID)
	if err != nil || org == nil {
		return c.Status(404).JSON(fiber.Map{"error": "organization not found"})
	}

	projectCount, _ := h.Store.GetOrgProjectCount(orgID)
	serviceCount, _ := h.Store.GetOrgServiceCount(orgID)
	deployCount, _ := h.Store.GetMonthlyDeploymentCount(orgID)

	return c.JSON(fiber.Map{
		"projects":    fiber.Map{"used": projectCount, "limit": org.MaxProjects},
		"services":    fiber.Map{"used": serviceCount, "limit": org.MaxServices},
		"deployments": fiber.Map{"used": deployCount, "limit": org.MaxDeployments},
		"cpu":         fiber.Map{"limit": org.MaxCPU},
		"memory":      fiber.Map{"limit": org.MaxMemory},
	})
}

// --- API Keys ---

func (h *Handler) ListAPIKeys(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	keys, err := h.Store.ListAPIKeys(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(keys)
}

func (h *Handler) CreateAPIKey(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var input model.CreateAPIKeyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	if input.Scopes == "" {
		input.Scopes = "*"
	}

	// Generate API key: mpk_live_<32 random hex chars>
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	rawKey := "mpk_live_" + hex.EncodeToString(randomBytes)
	keyHash := crypto.HashAPIKey(rawKey)
	keyPrefix := rawKey[:16] // "mpk_live_xxxx..."

	apiKey := &model.APIKey{
		ID:        store.NewID(),
		UserID:    userID,
		Name:      input.Name,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Scopes:    input.Scopes,
		CreatedAt: time.Now(),
	}

	if err := h.Store.CreateAPIKey(apiKey); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.audit(c, "create", "api_key", apiKey.ID, "name: "+input.Name)

	// Return the raw key only once — it cannot be retrieved later
	return c.Status(201).JSON(fiber.Map{
		"id":         apiKey.ID,
		"name":       apiKey.Name,
		"key":        rawKey,
		"key_prefix": keyPrefix,
		"scopes":     apiKey.Scopes,
		"created_at": apiKey.CreatedAt,
		"message":    "Save this key now — it will not be shown again",
	})
}

func (h *Handler) DeleteAPIKey(c *fiber.Ctx) error {
	if err := h.Store.DeleteAPIKey(c.Params("id")); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	h.audit(c, "delete", "api_key", c.Params("id"), "")
	return c.JSON(fiber.Map{"message": "API key deleted"})
}

// --- Notification Channels ---

func (h *Handler) ListNotificationChannels(c *fiber.Ctx) error {
	orgID := c.Query("org_id", "")
	channels, err := h.Store.ListNotificationChannels(orgID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(channels)
}

func (h *Handler) CreateNotificationChannel(c *fiber.Ctx) error {
	var input model.CreateNotificationChannelInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Name == "" || input.Type == "" || input.Config == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name, type, and config required"})
	}
	if input.Type != "webhook" && input.Type != "slack" && input.Type != "email" {
		return c.Status(400).JSON(fiber.Map{"error": "type must be webhook, slack, or email"})
	}

	ch := &model.NotificationChannel{
		ID:        store.NewID(),
		OrgID:     input.OrgID,
		Name:      input.Name,
		Type:      input.Type,
		Config:    input.Config,
		Enabled:   input.Enabled,
		CreatedAt: time.Now(),
	}

	if err := h.Store.CreateNotificationChannel(ch); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.audit(c, "create", "notification_channel", ch.ID, "type: "+ch.Type+" name: "+ch.Name)
	return c.Status(201).JSON(ch)
}

func (h *Handler) DeleteNotificationChannel(c *fiber.Ctx) error {
	if err := h.Store.DeleteNotificationChannel(c.Params("id")); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	h.audit(c, "delete", "notification_channel", c.Params("id"), "")
	return c.JSON(fiber.Map{"message": "notification channel deleted"})
}

func (h *Handler) UpdateNotificationChannelStatus(c *fiber.Ctx) error {
	var input struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := h.Store.UpdateNotificationChannel(c.Params("id"), input.Enabled); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "updated"})
}

// --- Notification Rules ---

func (h *Handler) ListNotificationRules(c *fiber.Ctx) error {
	rules, err := h.Store.ListNotificationRules(c.Params("channelId"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(rules)
}

func (h *Handler) CreateNotificationRule(c *fiber.Ctx) error {
	var input model.CreateNotificationRuleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Event == "" {
		return c.Status(400).JSON(fiber.Map{"error": "event is required"})
	}
	input.ChannelID = c.Params("channelId")

	rule := &model.NotificationRule{
		ID:        store.NewID(),
		ChannelID: input.ChannelID,
		Event:     input.Event,
		ProjectID: input.ProjectID,
		CreatedAt: time.Now(),
	}

	if err := h.Store.CreateNotificationRule(rule); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(rule)
}

func (h *Handler) DeleteNotificationRule(c *fiber.Ctx) error {
	if err := h.Store.DeleteNotificationRule(c.Params("ruleId")); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "rule deleted"})
}

// --- JWT Auth (enterprise) ---

func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil || body.RefreshToken == "" {
		return c.Status(400).JSON(fiber.Map{"error": "refresh_token is required"})
	}

	claims, err := crypto.ValidateJWT(body.RefreshToken, h.Config.Secret)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}
	if claims.Type != "refresh" {
		return c.Status(400).JSON(fiber.Map{"error": "must use refresh token"})
	}

	// Generate new access token
	accessToken, err := crypto.GenerateJWT(claims.Sub, claims.Username, claims.Role, "access", h.Config.Secret, h.Config.JWTExpiry)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"access_token": accessToken,
		"expires_in":   int(h.Config.JWTExpiry.Seconds()),
	})
}

// audit is a helper to log audit events.
func (h *Handler) audit(c *fiber.Ctx, action, resource, resourceID, details string) {
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)
	h.Store.AddAuditLog(userID, username, action, resource, resourceID, details)
}
