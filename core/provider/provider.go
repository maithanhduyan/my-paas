package provider

import (
	"github.com/my-paas/core/app"
	"github.com/my-paas/core/plan"
)

// Provider detects a language/framework and produces a build plan.
type Provider interface {
	Name() string
	Detect(a *app.App) (bool, error)
	Plan(a *app.App) (*plan.BuildPlan, error)
}
