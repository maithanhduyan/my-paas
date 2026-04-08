package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/model"
)

func (h *Handler) ListVolumes(c *fiber.Ctx) error {
	projectID := c.Params("id")
	volumes, err := h.Store.ListVolumes(projectID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if volumes == nil {
		volumes = []model.Volume{}
	}
	return c.JSON(volumes)
}

func (h *Handler) CreateVolume(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)

	var input model.CreateVolumeInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Name == "" || input.MountPath == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name and mount_path required"})
	}

	vol, err := h.Store.CreateVolume(projectID, input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.Store.AddAuditLog(userID, username, "create", "volume", vol.ID, "volume "+input.Name+" at "+input.MountPath)
	return c.Status(201).JSON(vol)
}

func (h *Handler) DeleteVolume(c *fiber.Ctx) error {
	id := c.Params("volumeId")
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)

	if err := h.Store.DeleteVolume(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.Store.AddAuditLog(userID, username, "delete", "volume", id, "volume deleted")
	return c.JSON(fiber.Map{"message": "volume deleted"})
}
