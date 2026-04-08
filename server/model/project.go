package model

import "time"

type Project struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	GitURL     string    `json:"git_url,omitempty"`
	Branch     string    `json:"branch"`
	Provider   string    `json:"provider,omitempty"`
	Framework  string    `json:"framework,omitempty"`
	AutoDeploy bool      `json:"auto_deploy"`
	Status     string    `json:"status"` // active | ready_to_deploy | deploying | healthy | failed
	CPULimit   float64   `json:"cpu_limit"`  // CPU cores (0 = unlimited)
	MemLimit   int64     `json:"mem_limit"`  // Memory in MB (0 = unlimited)
	Replicas   uint64    `json:"replicas"`   // Swarm replicas (0 = 1)
	CreatedBy  string    `json:"created_by"` // user ID
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateProjectInput struct {
	Name   string `json:"name"`
	GitURL string `json:"git_url,omitempty"`
	Branch string `json:"branch,omitempty"`
}

type UpdateProjectInput struct {
	Name       *string  `json:"name,omitempty"`
	Branch     *string  `json:"branch,omitempty"`
	AutoDeploy *bool    `json:"auto_deploy,omitempty"`
	CPULimit   *float64 `json:"cpu_limit,omitempty"`
	MemLimit   *int64   `json:"mem_limit,omitempty"`
	Replicas   *uint64  `json:"replicas,omitempty"`
}
