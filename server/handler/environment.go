package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/model"
)

func (h *Handler) GetEnvVars(c *fiber.Ctx) error {
	projectID := c.Params("id")

	vars, err := h.Store.GetEnvVars(projectID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if vars == nil {
		vars = []model.EnvVar{}
	}

	// Mask secret values in response
	for i := range vars {
		if vars[i].IsSecret {
			vars[i].Value = "********"
		}
	}

	return c.JSON(vars)
}

func (h *Handler) UpdateEnvVars(c *fiber.Ctx) error {
	projectID := c.Params("id")

	project, err := h.Store.GetProject(projectID)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	var input model.BulkEnvInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.Store.SetEnvVars(projectID, input.Vars); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Mark project as "ready to deploy" if it was healthy before
	if project.Status == "healthy" {
		h.Store.UpdateProjectStatus(projectID, "ready_to_deploy")
	}

	return c.JSON(fiber.Map{
		"message": "environment updated",
		"status":  "ready_to_deploy",
	})
}

func (h *Handler) DeleteEnvVar(c *fiber.Ctx) error {
	projectID := c.Params("id")
	key := c.Params("key")

	if err := h.Store.DeleteEnvVar(projectID, key); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	project, _ := h.Store.GetProject(projectID)
	if project != nil && project.Status == "healthy" {
		h.Store.UpdateProjectStatus(projectID, "ready_to_deploy")
	}

	return c.JSON(fiber.Map{"message": "env var deleted"})
}
