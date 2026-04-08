package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/model"
)

func (h *Handler) ListDomains(c *fiber.Ctx) error {
	projectID := c.Params("id")
	domains, err := h.Store.GetDomains(projectID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if domains == nil {
		domains = []model.Domain{}
	}
	return c.JSON(domains)
}

func (h *Handler) AddDomain(c *fiber.Ctx) error {
	projectID := c.Params("id")

	project, err := h.Store.GetProject(projectID)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	var input model.CreateDomainInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if input.Domain == "" {
		return c.Status(400).JSON(fiber.Map{"error": "domain is required"})
	}

	domain, err := h.Store.CreateDomain(projectID, input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(domain)
}

func (h *Handler) DeleteDomain(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.Store.DeleteDomain(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "domain deleted"})
}
