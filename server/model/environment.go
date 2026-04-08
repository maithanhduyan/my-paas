package model

import "time"

type EnvVar struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	IsSecret  bool      `json:"is_secret"`
	CreatedAt time.Time `json:"created_at"`
}

type EnvVarInput struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

type BulkEnvInput struct {
	Vars []EnvVarInput `json:"vars"`
}
