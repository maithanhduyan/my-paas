package handler

import (
	"github.com/gofiber/fiber/v2"

	core "github.com/my-paas/core"
	"github.com/my-paas/server/model"
)

func (h *Handler) ListProjects(c *fiber.Ctx) error {
	projects, err := h.Store.ListProjects()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if projects == nil {
		projects = []model.Project{}
	}
	return c.JSON(projects)
}

func (h *Handler) GetProject(c *fiber.Ctx) error {
	id := c.Params("id")
	project, err := h.Store.GetProject(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}
	return c.JSON(project)
}

func (h *Handler) CreateProject(c *fiber.Ctx) error {
	var input model.CreateProjectInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if input.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}

	// Check for duplicate project name
	existing, err := h.Store.ListProjects()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	for _, p := range existing {
		if p.Name == input.Name {
			return c.Status(409).JSON(fiber.Map{"error": "a project with this name already exists"})
		}
	}

	project, err := h.Store.CreateProject(input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Auto-detect if git URL provided — run detection in background
	// (actual detection happens during deploy)

	return c.Status(201).JSON(project)
}

func (h *Handler) UpdateProject(c *fiber.Ctx) error {
	id := c.Params("id")
	var input model.UpdateProjectInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	project, err := h.Store.UpdateProject(id, input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(project)
}

func (h *Handler) DeleteProject(c *fiber.Ctx) error {
	id := c.Params("id")

	project, err := h.Store.GetProject(id)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	// Stop any running container (Swarm uses dot separator)
	containerID, _ := h.Docker.FindContainerByName(c.Context(), "mypaas-"+project.Name+".")
	if containerID == "" {
		containerID, _ = h.Docker.FindContainerByName(c.Context(), "mypaas-"+project.Name+"-")
	}
	if containerID != "" {
		h.Docker.StopContainer(c.Context(), containerID)
	}

	if err := h.Store.DeleteProject(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "project deleted"})
}

// DetectProject runs auto-detection on a project's git repo without deploying.
func (h *Handler) DetectProject(c *fiber.Ctx) error {
	var input struct {
		Path string `json:"path"`
	}
	if err := c.BodyParser(&input); err != nil || input.Path == "" {
		return c.Status(400).JSON(fiber.Map{"error": "path is required"})
	}

	result := core.Detect(input.Path)
	return c.JSON(result)
}

// RestartProject force-updates the Swarm service to restart all tasks.
func (h *Handler) RestartProject(c *fiber.Ctx) error {
	id := c.Params("id")
	project, err := h.Store.GetProject(id)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	serviceName := "mypaas-" + project.Name
	svc, err := h.Docker.FindSwarmServiceByName(c.Context(), serviceName)
	if err != nil || svc == nil {
		return c.Status(404).JSON(fiber.Map{"error": "service not found"})
	}

	if err := h.Docker.UpdateSwarmService(c.Context(), svc.ID, "", 0, nil); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "service restarting"})
}

// StopProject scales the Swarm service to 0 replicas.
func (h *Handler) StopProject(c *fiber.Ctx) error {
	id := c.Params("id")
	project, err := h.Store.GetProject(id)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	serviceName := "mypaas-" + project.Name
	svc, err := h.Docker.FindSwarmServiceByName(c.Context(), serviceName)
	if err != nil || svc == nil {
		return c.Status(404).JSON(fiber.Map{"error": "service not found"})
	}

	if err := h.Docker.ScaleSwarmService(c.Context(), svc.ID, 0); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	h.Store.UpdateProjectStatus(id, "stopped")
	return c.JSON(fiber.Map{"message": "service stopped"})
}

// StartProject scales the Swarm service back to configured replicas.
func (h *Handler) StartProject(c *fiber.Ctx) error {
	id := c.Params("id")
	project, err := h.Store.GetProject(id)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	serviceName := "mypaas-" + project.Name
	svc, err := h.Docker.FindSwarmServiceByName(c.Context(), serviceName)
	if err != nil || svc == nil {
		return c.Status(404).JSON(fiber.Map{"error": "service not found"})
	}

	replicas := project.Replicas
	if replicas == 0 {
		replicas = 1
	}

	if err := h.Docker.ScaleSwarmService(c.Context(), svc.ID, replicas); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	h.Store.UpdateProjectStatus(id, "healthy")
	return c.JSON(fiber.Map{"message": "service started"})
}
