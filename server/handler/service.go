package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/docker"
	"github.com/my-paas/server/model"
)

const serviceNetwork = "mypaas-network"

func (h *Handler) ListServices(c *fiber.Ctx) error {
	services, err := h.Store.ListServices()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if services == nil {
		services = []model.Service{}
	}
	return c.JSON(services)
}

func (h *Handler) CreateService(c *fiber.Ctx) error {
	var input model.CreateServiceInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if input.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	if input.Type == "" {
		return c.Status(400).JSON(fiber.Map{"error": "type is required"})
	}

	svc, err := h.Store.CreateService(input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(svc)
}

func (h *Handler) DeleteService(c *fiber.Ctx) error {
	id := c.Params("id")

	svc, err := h.Store.GetService(id)
	if err != nil || svc == nil {
		return c.Status(404).JSON(fiber.Map{"error": "service not found"})
	}

	// Stop container if running
	if svc.ContainerID != "" {
		h.Docker.StopContainer(c.Context(), svc.ContainerID)
	}

	if err := h.Store.DeleteService(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "service deleted"})
}

func (h *Handler) StartService(c *fiber.Ctx) error {
	id := c.Params("id")

	svc, err := h.Store.GetService(id)
	if err != nil || svc == nil {
		return c.Status(404).JSON(fiber.Map{"error": "service not found"})
	}

	if svc.Status == "running" && svc.ContainerID != "" {
		return c.JSON(fiber.Map{"message": "service already running"})
	}

	// Get defaults for this service type
	_, defaultEnv, port := model.ServiceDefaults(svc.Type)

	containerName := "mypaas-svc-" + svc.Name

	// Stop old container if exists
	if svc.ContainerID != "" {
		h.Docker.StopContainer(c.Context(), svc.ContainerID)
	}
	// Also try by name
	oldID, _ := h.Docker.FindContainerByName(c.Context(), containerName)
	if oldID != "" {
		h.Docker.StopContainer(c.Context(), oldID)
	}

	// Pull image first
	if err := h.Docker.PullImage(c.Context(), svc.Image); err != nil {
		h.Store.UpdateServiceStatus(id, "error", "")
		return c.Status(500).JSON(fiber.Map{"error": "failed to pull image: " + err.Error()})
	}

	containerID, err := h.Docker.RunContainer(c.Context(), docker.RunContainerOpts{
		Name:    containerName,
		Image:   svc.Image,
		Env:     defaultEnv,
		Ports:   []string{port},
		Network: serviceNetwork,
		Labels: map[string]string{
			"mypaas.service": svc.ID,
			"mypaas.type":    string(svc.Type),
		},
	})
	if err != nil {
		h.Store.UpdateServiceStatus(id, "error", "")
		return c.Status(500).JSON(fiber.Map{"error": "failed to start: " + err.Error()})
	}

	h.Store.UpdateServiceStatus(id, "running", containerID)
	return c.JSON(fiber.Map{"message": "service started", "container_id": containerID})
}

func (h *Handler) StopService(c *fiber.Ctx) error {
	id := c.Params("id")

	svc, err := h.Store.GetService(id)
	if err != nil || svc == nil {
		return c.Status(404).JSON(fiber.Map{"error": "service not found"})
	}

	if svc.ContainerID != "" {
		h.Docker.StopContainer(c.Context(), svc.ContainerID)
	}

	h.Store.UpdateServiceStatus(id, "stopped", "")
	return c.JSON(fiber.Map{"message": "service stopped"})
}

func (h *Handler) LinkServiceToProject(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	projectID := c.Params("projectId")

	svc, err := h.Store.GetService(serviceID)
	if err != nil || svc == nil {
		return c.Status(404).JSON(fiber.Map{"error": "service not found"})
	}

	project, err := h.Store.GetProject(projectID)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	// Default env prefix based on service type
	envPrefix := "DATABASE_"
	if svc.Type == model.ServiceRedis {
		envPrefix = "REDIS_"
	} else if svc.Type == model.ServiceMongo {
		envPrefix = "MONGO_"
	} else if svc.Type == model.ServiceMinio {
		envPrefix = "S3_"
	}

	var input struct {
		EnvPrefix string `json:"env_prefix"`
	}
	if err := c.BodyParser(&input); err == nil && input.EnvPrefix != "" {
		envPrefix = input.EnvPrefix
	}

	link, err := h.Store.LinkService(projectID, serviceID, envPrefix)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Auto-inject connection env vars
	connEnv := model.ServiceConnectionEnv(svc, envPrefix)
	if len(connEnv) > 0 {
		var envInputs []model.EnvVarInput
		for k, v := range connEnv {
			envInputs = append(envInputs, model.EnvVarInput{Key: k, Value: v, IsSecret: false})
		}
		h.Store.SetEnvVars(projectID, envInputs)
	}

	return c.JSON(fiber.Map{"message": "service linked", "link": link, "env_injected": connEnv})
}

func (h *Handler) UnlinkServiceFromProject(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	projectID := c.Params("projectId")

	if err := h.Store.UnlinkService(projectID, serviceID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "service unlinked"})
}
