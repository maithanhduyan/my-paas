package php

import (
	"strings"

	"github.com/my-paas/core/app"
	"github.com/my-paas/core/plan"
)

type PhpProvider struct{}

func (p *PhpProvider) Name() string { return "php" }

func (p *PhpProvider) Detect(a *app.App) (bool, error) {
	return a.HasFile("composer.json") || a.HasFile("index.php"), nil
}

func (p *PhpProvider) Plan(a *app.App) (*plan.BuildPlan, error) {
	phpVersion := "8.3"
	framework := detectPhpFramework(a)

	bp := &plan.BuildPlan{
		Provider:   "php",
		Language:   "PHP",
		Version:    phpVersion,
		Framework:  framework,
		BaseImage:  "php:" + phpVersion + "-fpm-alpine",
		InstallCmd: "composer install --no-dev --optimize-autoloader",
		Ports:      []string{"8080"},
		EnvVars:    map[string]string{},
	}

	switch framework {
	case "laravel":
		bp.StartCmd = "php artisan serve --host=0.0.0.0 --port=8080"
		bp.BuildCmd = "php artisan config:cache && php artisan route:cache"
	case "symfony":
		bp.StartCmd = "php -S 0.0.0.0:8080 -t public"
	default:
		bp.StartCmd = "php -S 0.0.0.0:8080"
	}

	return bp, nil
}

func detectPhpFramework(a *app.App) string {
	if a.HasFile("artisan") {
		return "laravel"
	}
	if a.HasFile("composer.json") {
		content, err := a.ReadFile("composer.json")
		if err == nil {
			lower := strings.ToLower(content)
			if strings.Contains(lower, "laravel/framework") {
				return "laravel"
			}
			if strings.Contains(lower, "symfony/framework-bundle") {
				return "symfony"
			}
		}
	}
	return ""
}
