package staticfile

import (
	"github.com/my-paas/core/app"
	"github.com/my-paas/core/plan"
)

type StaticfileProvider struct{}

func (p *StaticfileProvider) Name() string { return "staticfile" }

func (p *StaticfileProvider) Detect(a *app.App) (bool, error) {
	return a.HasFile("index.html"), nil
}

func (p *StaticfileProvider) Plan(a *app.App) (*plan.BuildPlan, error) {
	return &plan.BuildPlan{
		Provider:  "staticfile",
		Language:  "Static",
		BaseImage: "nginx:alpine",
		StartCmd:  "nginx -g 'daemon off;'",
		Ports:     []string{"80"},
		Static:    true,
		EnvVars:   map[string]string{},
	}, nil
}
