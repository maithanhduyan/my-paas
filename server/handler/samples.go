package handler

import (
	"github.com/gofiber/fiber/v2"
)

type Sample struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Icon        string `json:"icon"`
	GitURL      string `json:"git_url"`
}

var builtinSamples = []Sample{
	{
		ID:          "node-app",
		Name:        "Node.js App",
		Description: "Express.js web server with health check",
		Language:    "node",
		Icon:        "⬢",
		GitURL:      "file:///data/samples/node-app.git",
	},
	{
		ID:          "python-app",
		Name:        "Python App",
		Description: "Flask web server with Gunicorn",
		Language:    "python",
		Icon:        "🐍",
		GitURL:      "file:///data/samples/python-app.git",
	},
	{
		ID:          "go-app",
		Name:        "Go App",
		Description: "Standard library HTTP server",
		Language:    "go",
		Icon:        "🐹",
		GitURL:      "file:///data/samples/go-app.git",
	},
	{
		ID:          "static-site",
		Name:        "Static Site",
		Description: "HTML/CSS static website",
		Language:    "static",
		Icon:        "🌐",
		GitURL:      "file:///data/samples/static-site.git",
	},
}

func (h *Handler) ListSamples(c *fiber.Ctx) error {
	return c.JSON(builtinSamples)
}
