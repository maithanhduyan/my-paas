package model

import "time"

type DeploymentStatus string

const (
	DeployQueued    DeploymentStatus = "queued"
	DeployCloning   DeploymentStatus = "cloning"
	DeployDetecting DeploymentStatus = "detecting"
	DeployBuilding  DeploymentStatus = "building"
	DeployDeploying DeploymentStatus = "deploying"
	DeployHealthy   DeploymentStatus = "healthy"
	DeployFailed    DeploymentStatus = "failed"
	DeployRolledBack DeploymentStatus = "rolled_back"
	DeployCancelled DeploymentStatus = "cancelled"
)

type DeployTrigger string

const (
	TriggerManual    DeployTrigger = "manual"
	TriggerWebhook   DeployTrigger = "webhook"
	TriggerEnvChange DeployTrigger = "env_change"
	TriggerRollback  DeployTrigger = "rollback"
)

type Deployment struct {
	ID         string           `json:"id"`
	ProjectID  string           `json:"project_id"`
	CommitHash string           `json:"commit_hash,omitempty"`
	CommitMsg  string           `json:"commit_msg,omitempty"`
	Status     DeploymentStatus `json:"status"`
	ImageTag   string           `json:"image_tag,omitempty"`
	Trigger    DeployTrigger    `json:"trigger"`
	StartedAt  *time.Time       `json:"started_at,omitempty"`
	FinishedAt *time.Time       `json:"finished_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
}

type DeploymentLog struct {
	ID           string    `json:"id"`
	DeploymentID string    `json:"deployment_id"`
	Step         string    `json:"step"`  // clone | detect | build | deploy | healthcheck
	Level        string    `json:"level"` // info | warn | error
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}
