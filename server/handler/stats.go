package handler

import (
	"context"
	"encoding/json"
	"io"

	"github.com/gofiber/fiber/v2"
)

type ContainerStats struct {
	Name      string  `json:"name"`
	ID        string  `json:"id"`
	CPUPerc   float64 `json:"cpu_percent"`
	MemUsage  uint64  `json:"mem_usage"`
	MemLimit  uint64  `json:"mem_limit"`
	MemPerc   float64 `json:"mem_percent"`
	NetInput  uint64  `json:"net_input"`
	NetOutput uint64  `json:"net_output"`
	Status    string  `json:"status"`
}

func (h *Handler) GetSystemStats(c *fiber.Ctx) error {
	containers, err := h.Docker.ListContainers(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var stats []ContainerStats
	for _, ctr := range containers {
		s, err := h.getContainerStats(c.Context(), ctr.ID)
		if err != nil {
			continue
		}
		name := ""
		if len(ctr.Names) > 0 {
			name = ctr.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}
		s.Name = name
		s.Status = ctr.State
		stats = append(stats, *s)
	}

	if stats == nil {
		stats = []ContainerStats{}
	}
	return c.JSON(stats)
}

func (h *Handler) GetProjectStats(c *fiber.Ctx) error {
	projectID := c.Params("id")
	project, err := h.Store.GetProject(projectID)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	containerID, _ := h.Docker.FindContainerByName(c.Context(), "mypaas-"+project.Name+".")
	if containerID == "" {
		containerID, _ = h.Docker.FindContainerByName(c.Context(), "mypaas-"+project.Name+"-")
	}
	if containerID == "" {
		return c.JSON(fiber.Map{"status": "no_container"})
	}

	s, err := h.getContainerStats(c.Context(), containerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	s.Name = "mypaas-" + project.Name
	return c.JSON(s)
}

func (h *Handler) getContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	resp, err := h.Docker.ContainerStatsOnce(ctx, containerID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw struct {
		CPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
			OnlineCPUs     uint32 `json:"online_cpus"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
		Networks map[string]struct {
			RxBytes uint64 `json:"rx_bytes"`
			TxBytes uint64 `json:"tx_bytes"`
		} `json:"networks"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	// Calculate CPU percentage
	cpuDelta := float64(raw.CPUStats.CPUUsage.TotalUsage - raw.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(raw.CPUStats.SystemCPUUsage - raw.PreCPUStats.SystemCPUUsage)
	cpuPerc := 0.0
	if sysDelta > 0 && cpuDelta > 0 {
		cpuPerc = (cpuDelta / sysDelta) * float64(raw.CPUStats.OnlineCPUs) * 100.0
	}

	// Memory percentage
	memPerc := 0.0
	if raw.MemoryStats.Limit > 0 {
		memPerc = float64(raw.MemoryStats.Usage) / float64(raw.MemoryStats.Limit) * 100.0
	}

	// Network IO
	var netIn, netOut uint64
	for _, net := range raw.Networks {
		netIn += net.RxBytes
		netOut += net.TxBytes
	}

	return &ContainerStats{
		ID:        containerID,
		CPUPerc:   cpuPerc,
		MemUsage:  raw.MemoryStats.Usage,
		MemLimit:  raw.MemoryStats.Limit,
		MemPerc:   memPerc,
		NetInput:  netIn,
		NetOutput: netOut,
	}, nil
}
