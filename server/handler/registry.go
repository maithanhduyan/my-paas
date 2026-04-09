package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	registryServiceName = "mypaas-registry"
	registryImage       = "registry:2"
	registryPort        = "5000"
	registryNetwork     = "mypaas-network"
)

type RegistryInfo struct {
	Status  string `json:"status"`
	Address string `json:"address"`
	Port    string `json:"port"`
}

type RegistryImage struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// GetRegistryStatus returns the status of the local Docker registry.
func (h *Handler) GetRegistryStatus(c *fiber.Ctx) error {
	ctx := c.Context()

	svc, _ := h.Docker.FindSwarmServiceByName(ctx, registryServiceName)
	if svc == nil {
		// Check non-swarm
		cid, _ := h.Docker.FindContainerByName(ctx, registryServiceName)
		if cid == "" {
			return c.JSON(RegistryInfo{Status: "not_running"})
		}
		status, _ := h.Docker.ContainerStatus(ctx, cid)
		return c.JSON(RegistryInfo{
			Status:  status,
			Address: "localhost",
			Port:    registryPort,
		})
	}

	return c.JSON(RegistryInfo{
		Status:  "running",
		Address: registryServiceName,
		Port:    registryPort,
	})
}

// StartRegistry starts a local Docker registry as a Swarm service or container.
func (h *Handler) StartRegistry(c *fiber.Ctx) error {
	ctx := c.Context()

	// Check if already running
	if h.Docker.IsSwarmActive(ctx) {
		existing, _ := h.Docker.FindSwarmServiceByName(ctx, registryServiceName)
		if existing != nil {
			return c.JSON(fiber.Map{"message": "registry already running", "address": registryServiceName + ":" + registryPort})
		}

		// Deploy as Swarm service with published port
		_, err := h.Docker.CreateSwarmRegistryService(ctx, registryServiceName, registryImage, registryPort, registryNetwork)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to start registry: " + err.Error()})
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "registry started",
			"address": registryServiceName + ":" + registryPort,
		})
	}

	// Non-swarm: run as container
	cid, _ := h.Docker.FindContainerByName(ctx, registryServiceName)
	if cid != "" {
		return c.JSON(fiber.Map{"message": "registry already running", "address": "localhost:" + registryPort})
	}

	cid, err := h.Docker.RunRegistryContainer(ctx, registryServiceName, registryImage, registryPort)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to start registry: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":      "registry started",
		"address":      "localhost:" + registryPort,
		"container_id": cid,
	})
}

// StopRegistry stops the local Docker registry.
func (h *Handler) StopRegistry(c *fiber.Ctx) error {
	ctx := c.Context()

	if h.Docker.IsSwarmActive(ctx) {
		svc, _ := h.Docker.FindSwarmServiceByName(ctx, registryServiceName)
		if svc != nil {
			h.Docker.RemoveSwarmService(ctx, svc.ID)
		}
	} else {
		cid, _ := h.Docker.FindContainerByName(ctx, registryServiceName)
		if cid != "" {
			h.Docker.StopContainer(ctx, cid)
		}
	}

	return c.JSON(fiber.Map{"message": "registry stopped"})
}

// ListRegistryImages lists images in the local Docker registry.
func (h *Handler) ListRegistryImages(c *fiber.Ctx) error {
	registryURL := h.getRegistryURL(c.Context())
	if registryURL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "registry is not running"})
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(registryURL + "/v2/_catalog")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to query registry: " + err.Error()})
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var catalog struct {
		Repositories []string `json:"repositories"`
	}
	json.Unmarshal(data, &catalog)

	images := make([]RegistryImage, 0, len(catalog.Repositories))
	for _, repo := range catalog.Repositories {
		tags := h.getRegistryTags(client, registryURL, repo)
		images = append(images, RegistryImage{Name: repo, Tags: tags})
	}

	return c.JSON(images)
}

// DeleteRegistryImage deletes an image tag from the local registry.
func (h *Handler) DeleteRegistryImage(c *fiber.Ctx) error {
	name := c.Params("name")
	tag := c.Query("tag", "latest")

	registryURL := h.getRegistryURL(c.Context())
	if registryURL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "registry is not running"})
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Get manifest digest
	req, _ := http.NewRequest("GET", registryURL+"/v2/"+name+"/manifests/"+tag, nil)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return c.Status(404).JSON(fiber.Map{"error": "image not found"})
	}

	// Delete by digest
	delReq, _ := http.NewRequest("DELETE", registryURL+"/v2/"+name+"/manifests/"+digest, nil)
	delResp, err := client.Do(delReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer delResp.Body.Close()

	if delResp.StatusCode != 202 {
		body, _ := io.ReadAll(delResp.Body)
		return c.Status(delResp.StatusCode).JSON(fiber.Map{"error": string(body)})
	}

	return c.JSON(fiber.Map{"message": fmt.Sprintf("deleted %s:%s", name, tag)})
}

// PushToRegistry tags and pushes a project image to the local registry.
func (h *Handler) PushToRegistry(c *fiber.Ctx) error {
	projectID := c.Params("id")
	project, err := h.Store.GetProject(projectID)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	ctx := c.Context()
	registryAddr := h.getRegistryAddr(ctx)
	if registryAddr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "registry is not running"})
	}

	sourceImage := "mypaas-" + project.Name + ":latest"
	targetImage := registryAddr + "/" + project.Name + ":latest"

	// Tag image
	if err := h.Docker.TagImage(ctx, sourceImage, targetImage); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to tag image: " + err.Error()})
	}

	// Push image
	if err := h.Docker.PushImage(ctx, targetImage); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to push image: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "pushed to registry",
		"image":   targetImage,
	})
}

func (h *Handler) getRegistryURL(ctx context.Context) string {
	// For HTTP API calls from the server container, use service DNS when in Swarm
	if h.Docker.IsSwarmActive(ctx) {
		svc, _ := h.Docker.FindSwarmServiceByName(ctx, registryServiceName)
		if svc != nil {
			return "http://" + registryServiceName + ":" + registryPort
		}
	} else {
		cid, _ := h.Docker.FindContainerByName(ctx, registryServiceName)
		if cid != "" {
			return "http://localhost:" + registryPort
		}
	}
	return ""
}

func (h *Handler) getRegistryAddr(ctx context.Context) string {
	if h.Docker.IsSwarmActive(ctx) {
		svc, _ := h.Docker.FindSwarmServiceByName(ctx, registryServiceName)
		if svc != nil {
			// Use localhost because docker push happens at daemon level,
			// not inside the overlay network where service DNS works.
			return "localhost:" + registryPort
		}
	} else {
		cid, _ := h.Docker.FindContainerByName(ctx, registryServiceName)
		if cid != "" {
			return "localhost:" + registryPort
		}
	}
	return ""
}

func (h *Handler) getRegistryTags(client *http.Client, registryURL, repo string) []string {
	resp, err := client.Get(registryURL + "/v2/" + repo + "/tags/list")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var tagResp struct {
		Tags []string `json:"tags"`
	}
	json.Unmarshal(data, &tagResp)

	// Filter out empty
	tags := make([]string, 0, len(tagResp.Tags))
	for _, t := range tagResp.Tags {
		if strings.TrimSpace(t) != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
