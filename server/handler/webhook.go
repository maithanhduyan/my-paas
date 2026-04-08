package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/model"
	"github.com/my-paas/server/worker"
)

type githubPushPayload struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	HeadCommit *struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	} `json:"head_commit"`
	Repository struct {
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func (h *Handler) GithubWebhook(c *fiber.Ctx) error {
	// Verify webhook signature if secret is configured
	webhookSecret := os.Getenv("MYPAAS_WEBHOOK_SECRET")
	if webhookSecret != "" {
		sig := c.Get("X-Hub-Signature-256")
		if sig == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing signature"})
		}
		if !verifyGithubSignature(c.Body(), sig, webhookSecret) {
			return c.Status(401).JSON(fiber.Map{"error": "invalid signature"})
		}
	}

	// Only handle push events
	event := c.Get("X-GitHub-Event")
	if event == "ping" {
		return c.JSON(fiber.Map{"message": "pong"})
	}
	if event != "push" {
		return c.JSON(fiber.Map{"message": "ignored event: " + event})
	}

	var payload githubPushPayload
	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	// Extract branch from ref (refs/heads/main → main)
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Find matching project by git URL
	projects, err := h.Store.ListProjects()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var matched *model.Project
	for i := range projects {
		p := &projects[i]
		// Match by clone URL or HTML URL
		if normalizeGitURL(p.GitURL) == normalizeGitURL(payload.Repository.CloneURL) ||
			normalizeGitURL(p.GitURL) == normalizeGitURL(payload.Repository.HTMLURL) {
			if p.Branch == branch && p.AutoDeploy {
				matched = p
				break
			}
		}
	}

	if matched == nil {
		return c.JSON(fiber.Map{"message": "no matching project found"})
	}

	// Create deployment and queue it
	deploy, err := h.Store.CreateDeployment(matched.ID, model.TriggerWebhook)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Pre-set commit info from webhook
	if payload.HeadCommit != nil {
		shortHash := payload.HeadCommit.ID
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		h.Store.UpdateDeploymentCommit(deploy.ID, shortHash, payload.HeadCommit.Message)
	}

	err = h.Queue.Enqueue(worker.Job{
		Type:         worker.JobDeploy,
		ProjectID:    matched.ID,
		DeploymentID: deploy.ID,
	})
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "queue full"})
	}

	log.Printf("[webhook] auto-deploy triggered for project=%s (%s) branch=%s", matched.ID, matched.Name, branch)

	return c.Status(202).JSON(fiber.Map{
		"message":       "deploy triggered",
		"project_id":    matched.ID,
		"deployment_id": deploy.ID,
		"branch":        branch,
	})
}

func verifyGithubSignature(payload []byte, signature, secret string) bool {
	sig := strings.TrimPrefix(signature, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

func normalizeGitURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimSuffix(url, "/")
	url = strings.Replace(url, "git@github.com:", "https://github.com/", 1)
	return strings.ToLower(url)
}
