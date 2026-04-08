package model

type Template struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Icon        string            `json:"icon"`
	Services    []TemplateService `json:"services"`
}

type TemplateService struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`    // app | postgres | redis | mysql | mongo | minio
	GitURL  string            `json:"git_url"` // for app type
	Image   string            `json:"image"`   // for service type
	Env     map[string]string `json:"env"`
	Volumes []string          `json:"volumes"` // mount paths
}
