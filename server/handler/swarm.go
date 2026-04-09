package handler

import (
	"github.com/docker/docker/api/types/swarm"
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

		// Include manager advertise address
		managerAddr, _ := h.Docker.SwarmManagerAddr(ctx)
		result["manager_addr"] = managerAddr
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

// SwarmServices returns all Swarm services with their tasks.
func (h *Handler) SwarmServices(c *fiber.Ctx) error {
	ctx := c.Context()
	if !h.Docker.IsSwarmActive(ctx) {
		return c.JSON([]interface{}{})
	}

	services, err := h.Docker.ListSwarmServices(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Build node map for hostname lookup
	nodes, _ := h.Docker.ListSwarmNodes(ctx)
	nodeMap := make(map[string]string) // nodeID -> hostname
	for _, n := range nodes {
		nodeMap[n.ID] = n.Description.Hostname
	}

	result := make([]fiber.Map, 0, len(services))
	for _, svc := range services {
		var replicas uint64
		if svc.Spec.Mode.Replicated != nil && svc.Spec.Mode.Replicated.Replicas != nil {
			replicas = *svc.Spec.Mode.Replicated.Replicas
		}

		image := ""
		if svc.Spec.TaskTemplate.ContainerSpec != nil {
			image = svc.Spec.TaskTemplate.ContainerSpec.Image
		}

		// Get tasks for this service
		tasks, _ := h.Docker.ListSwarmServiceTasks(ctx, svc.ID)
		taskList := make([]fiber.Map, 0, len(tasks))
		for _, t := range tasks {
			if t.Status.State == swarm.TaskStateRunning ||
				t.Status.State == swarm.TaskStatePending ||
				t.Status.State == swarm.TaskStateStarting ||
				t.Status.State == swarm.TaskStatePreparing {
				taskList = append(taskList, fiber.Map{
					"id":        t.ID,
					"node_id":   t.NodeID,
					"node_name": nodeMap[t.NodeID],
					"state":     string(t.Status.State),
					"message":   t.Status.Message,
				})
			}
		}

		svcMap := fiber.Map{
			"id":       svc.ID,
			"name":     svc.Spec.Name,
			"image":    image,
			"replicas": replicas,
			"tasks":    taskList,
			"labels":   svc.Spec.Labels,
		}
		result = append(result, svcMap)
	}

	return c.JSON(result)
}
