package rust

import (
	"regexp"

	"github.com/my-paas/core/app"
	"github.com/my-paas/core/plan"
)

type RustProvider struct{}

func (p *RustProvider) Name() string { return "rust" }

func (p *RustProvider) Detect(a *app.App) (bool, error) {
	return a.HasFile("Cargo.toml"), nil
}

func (p *RustProvider) Plan(a *app.App) (*plan.BuildPlan, error) {
	bp := &plan.BuildPlan{
		Provider:   "rust",
		Language:   "Rust",
		Version:    "latest",
		BaseImage:  "rust:slim",
		InstallCmd: "",
		BuildCmd:   "cargo build --release",
		StartCmd:   "./target/release/app",
		Ports:      []string{"8080"},
		EnvVars:    map[string]string{},
	}

	if a.HasFile("Cargo.toml") {
		if content, err := a.ReadFile("Cargo.toml"); err == nil {
			if name := parseCrateNameFromCargo(content); name != "" {
				bp.StartCmd = "./target/release/" + name
			}
		}
	}

	return bp, nil
}

func parseCrateNameFromCargo(cargo string) string {
	re := regexp.MustCompile(`(?m)^name\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(cargo)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
