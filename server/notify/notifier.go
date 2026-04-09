package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/my-paas/server/model"
	"github.com/my-paas/server/store"
)

type Notifier struct {
	Store  *store.Store
	client *http.Client
}

func New(s *store.Store) *Notifier {
	return &Notifier{
		Store: s,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type EventPayload struct {
	Event     string      `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Project   *EventProject `json:"project,omitempty"`
	Deploy    *EventDeploy  `json:"deploy,omitempty"`
	Message   string      `json:"message,omitempty"`
}

type EventProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type EventDeploy struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	CommitHash string `json:"commit_hash,omitempty"`
	CommitMsg  string `json:"commit_msg,omitempty"`
	Trigger    string `json:"trigger"`
}

// Send dispatches notifications for an event to all matching channels.
func (n *Notifier) Send(ctx context.Context, event string, payload EventPayload) {
	projectID := ""
	if payload.Project != nil {
		projectID = payload.Project.ID
	}
	payload.Event = event
	payload.Timestamp = time.Now()

	channels, err := n.Store.GetNotificationRulesForEvent(event, projectID)
	if err != nil {
		log.Printf("[notify] failed to get channels for event %s: %v", event, err)
		return
	}

	for _, ch := range channels {
		go func(ch model.NotificationChannel) {
			var sendErr error
			switch ch.Type {
			case "webhook":
				sendErr = n.sendWebhook(ctx, ch, payload)
			case "slack":
				sendErr = n.sendSlack(ctx, ch, payload)
			case "email":
				log.Printf("[notify] email channel %s: not yet implemented", ch.Name)
				return
			default:
				log.Printf("[notify] unknown channel type: %s", ch.Type)
				return
			}
			if sendErr != nil {
				log.Printf("[notify] failed to send %s to channel %s: %v", event, ch.Name, sendErr)
			}
		}(ch)
	}
}

type webhookConfig struct {
	URL    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

func (n *Notifier) sendWebhook(ctx context.Context, ch model.NotificationChannel, payload EventPayload) error {
	var cfg webhookConfig
	if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
		return fmt.Errorf("parse webhook config: %w", err)
	}
	if cfg.URL == "" {
		return fmt.Errorf("webhook URL is empty")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MyPaaS/4.0")
	if cfg.Secret != "" {
		req.Header.Set("X-MyPaaS-Secret", cfg.Secret)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

type slackConfig struct {
	WebhookURL string `json:"webhook_url"`
}

func (n *Notifier) sendSlack(ctx context.Context, ch model.NotificationChannel, payload EventPayload) error {
	var cfg slackConfig
	if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
		return fmt.Errorf("parse slack config: %w", err)
	}
	if cfg.WebhookURL == "" {
		return fmt.Errorf("slack webhook URL is empty")
	}

	// Build Slack message
	text := formatSlackMessage(payload)

	slackPayload := map[string]interface{}{
		"text": text,
		"blocks": []map[string]interface{}{
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": text,
				},
			},
		},
	}

	body, err := json.Marshal(slackPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}
	return nil
}

func formatSlackMessage(p EventPayload) string {
	projectName := "unknown"
	if p.Project != nil {
		projectName = p.Project.Name
	}

	switch p.Event {
	case model.EventDeployStarted:
		return fmt.Sprintf(":rocket: Deploy started for *%s*", projectName)
	case model.EventDeploySucceeded:
		msg := fmt.Sprintf(":white_check_mark: Deploy succeeded for *%s*", projectName)
		if p.Deploy != nil && p.Deploy.CommitMsg != "" {
			msg += fmt.Sprintf("\n> %s", p.Deploy.CommitMsg)
		}
		return msg
	case model.EventDeployFailed:
		return fmt.Sprintf(":x: Deploy failed for *%s*", projectName)
	case model.EventHealthDown:
		return fmt.Sprintf(":warning: *%s* is down!", projectName)
	case model.EventHealthRecovered:
		return fmt.Sprintf(":green_heart: *%s* recovered", projectName)
	case model.EventQuotaWarning:
		return fmt.Sprintf(":warning: Quota nearing limit: %s", p.Message)
	case model.EventBackupCompleted:
		return fmt.Sprintf(":floppy_disk: Backup completed: %s", p.Message)
	default:
		return fmt.Sprintf("[%s] %s: %s", p.Event, projectName, p.Message)
	}
}
