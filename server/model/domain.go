package model

import "time"

type Domain struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Domain    string    `json:"domain"`
	SSLAuto   bool      `json:"ssl_auto"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateDomainInput struct {
	Domain  string `json:"domain"`
	SSLAuto *bool  `json:"ssl_auto,omitempty"`
}
