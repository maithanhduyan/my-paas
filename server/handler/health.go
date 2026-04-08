package handler

import (
	"runtime"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	err := h.Docker.Ping(c.Context())
	dockerStatus := "connected"
	if err != nil {
		dockerStatus = "disconnected: " + err.Error()
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"docker": dockerStatus,
		"go":     runtime.Version(),
	})
}
