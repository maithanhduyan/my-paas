package model

import "time"

type AuditLog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`    // create | update | delete | deploy | rollback | login | logout | backup | restore
	Resource  string    `json:"resource"`  // project | service | domain | env | user | system
	ResourceID string   `json:"resource_id"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

type Invitation struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Token     string    `json:"token"`
	Used      bool      `json:"used"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
