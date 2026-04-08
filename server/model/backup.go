package model

import "time"

type Backup struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`      // system | service
	ServiceID string    `json:"service_id"` // empty for system backups
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}
