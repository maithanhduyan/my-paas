package model

import "time"

type Volume struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	MountPath string    `json:"mount_path"` // e.g. /data, /uploads
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateVolumeInput struct {
	Name      string `json:"name"`
	MountPath string `json:"mount_path"`
}
