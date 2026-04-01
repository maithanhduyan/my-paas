package golang

import (
	"regexp"
	"strings"

	"github.com/my-paas/core/app"
	"github.com/my-paas/core/plan"
)

type GoProvider struct{}

func (p *GoProvider) Name() string { return "golang" }

func (p *GoProvider) Detect(a *app.App) (bool, error) {
	return a.HasFile("go.mod") || a.HasFile("main.go"), nil
}

func (p *GoProvider) Plan(a *app.App) (*plan.BuildPlan, error) {
	goVersion := "1.22"
	if a.HasFile("go.mod") {
		if content, err := a.ReadFile("go.mod"); err == nil {
			if v := parseGoVersion(content); v != "" {
				goVersion = v
			}
		}
	}

	framework := detectGoFramework(a)

	bp := &plan.BuildPlan{
		Provider:   "golang",
		Language:   "Go",
		Version:    goVersion,
		Framework:  framework,
		BaseImage:  "golang:" + goVersion + "-alpine",
		InstallCmd: "go mod download",
		BuildCmd:   "CGO_ENABLED=0 go build -ldflags=\"-w -s\" -o /app/server .",
		StartCmd:   "./server",
		Ports:      []string{"8080"},
		EnvVars:    map[string]string{},
	}

	return bp, nil
}

func parseGoVersion(gomod string) string {
	re := regexp.MustCompile(`(?m)^go\s+(\d+\.\d+)`)
	matches := re.FindStringSubmatch(gomod)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func detectGoFramework(a *app.App) string {
	if !a.HasFile("go.mod") {
		return ""
	}
	content, err := a.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	if strings.Contains(content, "github.com/gin-gonic/gin") {
		return "gin"
	}
	if strings.Contains(content, "github.com/labstack/echo") {
		return "echo"
	}
	if strings.Contains(content, "github.com/gofiber/fiber") {
		return "fiber"
	}
	if strings.Contains(content, "github.com/gorilla/mux") {
		return "gorilla"
	}
	return ""
}
