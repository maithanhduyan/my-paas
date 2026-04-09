package handler

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// StreamProjectLogs streams container logs for a running project via SSE.
func (h *Handler) StreamProjectLogs(c *fiber.Ctx) error {
	projectID := c.Params("id")

	project, err := h.Store.GetProject(projectID)
	if err != nil || project == nil {
		return c.Status(404).JSON(fiber.Map{"error": "project not found"})
	}

	// Find running container for this project
	containerID, _ := h.Docker.FindContainerByName(c.Context(), "mypaas-"+project.Name+".")
	if containerID == "" {
		// Fallback: try with dash separator for non-swarm containers
		containerID, _ = h.Docker.FindContainerByName(c.Context(), "mypaas-"+project.Name+"-")
	}
	if containerID == "" {
		return c.Status(404).JSON(fiber.Map{"error": "no running container found"})
	}

	tail := c.Query("tail", "100")

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ctx := context.Background()
		reader, err := h.Docker.GetContainerLogs(ctx, containerID, true, tail)
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
			w.Flush()
			return
		}
		defer reader.Close()

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Bytes()
			// Docker multiplexed log lines have an 8-byte header; strip it
			if len(line) > 8 && (line[0] == 1 || line[0] == 2) && line[1] == 0 && line[2] == 0 && line[3] == 0 {
				line = line[8:]
			}
			text := strings.TrimSpace(string(line))
			if text == "" {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", text)
			w.Flush()
		}
	})

	return nil
}

// StreamDeploymentLogs streams build/deploy logs for a deployment via SSE.
func (h *Handler) StreamDeploymentLogs(c *fiber.Ctx) error {
	deployID := c.Params("id")

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		lastCount := 0
		for {
			logs, err := h.Store.GetDeploymentLogs(deployID)
			if err != nil {
				fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
				w.Flush()
				return
			}

			// Send only new logs
			for i := lastCount; i < len(logs); i++ {
				fmt.Fprintf(w, "data: {\"step\":\"%s\",\"level\":\"%s\",\"message\":\"%s\",\"time\":\"%s\"}\n\n",
					logs[i].Step, logs[i].Level, escapeJSON(logs[i].Message), logs[i].CreatedAt.Format(time.RFC3339))
				w.Flush()
			}
			lastCount = len(logs)

			// Check if deployment is finished
			deploy, _ := h.Store.GetDeployment(deployID)
			if deploy != nil {
				switch deploy.Status {
				case "healthy", "failed", "rolled_back", "cancelled":
					fmt.Fprintf(w, "data: {\"status\":\"%s\",\"done\":true}\n\n", deploy.Status)
					w.Flush()
					return
				}
			}

			time.Sleep(1 * time.Second)
		}
	})

	return nil
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", ``)
	return s
}
