package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/model"
	"github.com/my-paas/server/worker"
)

func (h *Handler) TriggerDeploy(c *fiber.Ctx) error {
	projectID := c.Params("id")

	project, err := h.Store.GetProject(projectID)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}
	if project.GitURL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "project has no git URL configured"})
	}

	deploy, err := h.Store.CreateDeployment(projectID, model.TriggerManual)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	err = h.Queue.Enqueue(worker.Job{
		Type:         worker.JobDeploy,
		ProjectID:    projectID,
		DeploymentID: deploy.ID,
	})
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "queue full, try again later"})
	}

	return c.Status(202).JSON(deploy)
}

func (h *Handler) ListDeployments(c *fiber.Ctx) error {
	projectID := c.Params("id")
	deployments, err := h.Store.ListDeployments(projectID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if deployments == nil {
		deployments = []model.Deployment{}
	}
	return c.JSON(deployments)
}

func (h *Handler) GetDeployment(c *fiber.Ctx) error {
	id := c.Params("id")
	deploy, err := h.Store.GetDeployment(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if deploy == nil {
		return c.Status(404).JSON(fiber.Map{"error": "deployment not found"})
	}
	return c.JSON(deploy)
}

func (h *Handler) RollbackDeployment(c *fiber.Ctx) error {
	id := c.Params("id")

	deploy, err := h.Store.GetDeployment(id)
	if err != nil || deploy == nil {
		return c.Status(404).JSON(fiber.Map{"error": "deployment not found"})
	}
	if deploy.ImageTag == "" {
		return c.Status(400).JSON(fiber.Map{"error": "deployment has no image to rollback to"})
	}

	err = h.Queue.Enqueue(worker.Job{
		Type:         worker.JobRollback,
		ProjectID:    deploy.ProjectID,
		DeploymentID: deploy.ID,
	})
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "queue full"})
	}

	return c.Status(202).JSON(fiber.Map{"message": "rollback queued"})
}

func (h *Handler) GetDeploymentLogs(c *fiber.Ctx) error {
	id := c.Params("id")
	logs, err := h.Store.GetDeploymentLogs(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if logs == nil {
		logs = []model.DeploymentLog{}
	}
	return c.JSON(logs)
}
