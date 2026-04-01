package core

import (
	"fmt"

	"github.com/my-paas/core/app"
	"github.com/my-paas/core/generate"
	"github.com/my-paas/core/plan"
	"github.com/my-paas/core/provider"
)

// DetectResult is the JSON output of the detect command.
type DetectResult struct {
	Success  bool            `json:"success"`
	Plan     *plan.BuildPlan `json:"plan,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// Detect analyzes a source directory and returns a build plan.
func Detect(path string) *DetectResult {
	a, err := app.NewApp(path)
	if err != nil {
		return &DetectResult{Success: false, Error: err.Error()}
	}

	providers := provider.GetProviders()

	for _, p := range providers {
		detected, err := p.Detect(a)
		if err != nil {
			continue
		}
		if detected {
			bp, err := p.Plan(a)
			if err != nil {
				return &DetectResult{Success: false, Error: fmt.Sprintf("provider %s error: %s", p.Name(), err.Error())}
			}
			return &DetectResult{Success: true, Plan: bp}
		}
	}

	return &DetectResult{Success: false, Error: "no provider matched the source directory"}
}

// GenerateDockerfile detects the app and generates a Dockerfile.
func GenerateDockerfile(path string) (string, error) {
	result := Detect(path)
	if !result.Success {
		return "", fmt.Errorf(result.Error)
	}
	return generate.GenerateDockerfile(result.Plan), nil
}
