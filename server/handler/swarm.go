package handler

import (
	"github.com/gofiber/fiber/v2"
)

// SwarmStatus returns the current Swarm state and node list.
func (h *Handler) SwarmStatus(c *fiber.Ctx) error {
	ctx := c.Context()
	active := h.Docker.IsSwarmActive(ctx)

	result := fiber.Map{
		"active": active,
		"nodes":  []interface{}{},
	}

	if active {
		nodes, err := h.Docker.ListSwarmNodes(ctx)
		if err == nil {
			nodeList := make([]fiber.Map, 0, len(nodes))
			for _, n := range nodes {
				nodeList = append(nodeList, fiber.Map{
					"id":       n.ID,
					"hostname": n.Description.Hostname,
					"status":   string(n.Status.State),
					"role":     string(n.Spec.Role),
					"addr":     n.Status.Addr,
				})
			}
			result["nodes"] = nodeList
		}
	}

	return c.JSON(result)
}

// SwarmInit initializes Docker Swarm on this node.
func (h *Handler) SwarmInit(c *fiber.Ctx) error {
	var body struct {
		AdvertiseAddr string `json:"advertise_addr"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if body.AdvertiseAddr == "" {
		body.AdvertiseAddr = "0.0.0.0:2377"
	}

	_, err := h.Docker.SwarmInit(c.Context(), body.AdvertiseAddr)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Swarm initialized"})
}

// SwarmToken returns the worker join token.
func (h *Handler) SwarmToken(c *fiber.Ctx) error {
	token, err := h.Docker.SwarmJoinToken(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"token": token})
}
