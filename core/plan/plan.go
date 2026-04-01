package plan

// BuildPlan is the output of analyzing an app source.
type BuildPlan struct {
	Provider   string            `json:"provider"`
	Language   string            `json:"language"`
	Version    string            `json:"version,omitempty"`
	Framework  string            `json:"framework,omitempty"`
	BaseImage  string            `json:"baseImage"`
	InstallCmd string            `json:"installCmd,omitempty"`
	BuildCmd   string            `json:"buildCmd,omitempty"`
	StartCmd   string            `json:"startCmd"`
	Ports      []string          `json:"ports"`
	EnvVars    map[string]string `json:"envVars,omitempty"`
	Paths      []string          `json:"paths,omitempty"`
	Caches     []string          `json:"caches,omitempty"`
	Static     bool              `json:"static,omitempty"`
}
