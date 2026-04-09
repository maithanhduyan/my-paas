package model

import "time"

// Organization represents a team/company in the system.
type Organization struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	MaxProjects    int       `json:"max_projects"`
	MaxServices    int       `json:"max_services"`
	MaxCPU         float64   `json:"max_cpu"`
	MaxMemory      int64     `json:"max_memory"`
	MaxDeployments int       `json:"max_deployments"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateOrgInput struct {
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	MaxProjects    int     `json:"max_projects"`
	MaxServices    int     `json:"max_services"`
	MaxCPU         float64 `json:"max_cpu"`
	MaxMemory      int64   `json:"max_memory"`
	MaxDeployments int     `json:"max_deployments"`
}

type UpdateOrgInput struct {
	Name           *string  `json:"name,omitempty"`
	MaxProjects    *int     `json:"max_projects,omitempty"`
	MaxServices    *int     `json:"max_services,omitempty"`
	MaxCPU         *float64 `json:"max_cpu,omitempty"`
	MaxMemory      *int64   `json:"max_memory,omitempty"`
	MaxDeployments *int     `json:"max_deployments,omitempty"`
}

type OrgMember struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"` // owner | admin | member | viewer
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type AddOrgMemberInput struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// APIKey represents a long-lived API key for CI/CD integrations.
type APIKey struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      string     `json:"name"`
	KeyHash   string     `json:"-"`
	KeyPrefix string     `json:"key_prefix"` // e.g. "mpk_live_ab12"
	Scopes    string     `json:"scopes"`     // comma-separated: "projects:read,deployments:write" or "*"
	LastUsed  *time.Time `json:"last_used,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateAPIKeyInput struct {
	Name   string `json:"name"`
	Scopes string `json:"scopes,omitempty"` // defaults to "*"
}

// NotificationChannel represents a configured alert destination.
type NotificationChannel struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`   // webhook | slack | email
	Config    string    `json:"config"` // JSON: {url, secret} or {webhook_url} or {smtp_host, from, to}
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateNotificationChannelInput struct {
	OrgID   string `json:"org_id,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Config  string `json:"config"`
	Enabled bool   `json:"enabled"`
}

// NotificationRule links events to channels.
type NotificationRule struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	Event     string    `json:"event"`      // deploy.started, deploy.succeeded, deploy.failed, health.down, etc.
	ProjectID string    `json:"project_id"` // empty = all projects
	CreatedAt time.Time `json:"created_at"`
}

type CreateNotificationRuleInput struct {
	ChannelID string `json:"channel_id"`
	Event     string `json:"event"`
	ProjectID string `json:"project_id,omitempty"`
}

// Notification events
const (
	EventDeployStarted   = "deploy.started"
	EventDeploySucceeded = "deploy.succeeded"
	EventDeployFailed    = "deploy.failed"
	EventHealthDown      = "health.down"
	EventHealthRecovered = "health.recovered"
	EventQuotaWarning    = "quota.warning"
	EventBackupCompleted = "backup.completed"
)
